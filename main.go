// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hekmon/transmissionrpc/v2"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type TemplateArgs struct {
	Torrents       []Torrent
	Error          error
	Request        *http.Request
	VAPIDPublicKey string
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon"`
}

const (
	vapidFilename    = "vapid.json"
	torrentsFilename = "torrents.json"
)

var (
	rpcHost string
	rpcUser string
	rpcPass string

	rpcTLS     = false
	rpcPort    = 9091
	listenPort = 8080
	baseURI    = "http://localhost:8080"

	vapid    VAPIDKeyPair
	torrents = NewTorrentStore()

	tpl    *template.Template
	client *transmissionrpc.Client
)

func handleAddTorrent(r *http.Request, session string) error {
	if err := r.ParseMultipartForm(200000); err != nil {
		return err
	}

	newTorrents := []transmissionrpc.Torrent{}

	var push *webpush.Subscription
	if ss, ok := r.MultipartForm.Value["subscription"]; ok && len(ss) == 1 {
		if sub := ss[0]; sub != "" {
			push = &webpush.Subscription{}
			if err := json.Unmarshal([]byte(sub), push); err != nil {
				return fmt.Errorf("failed to decode WebPush subscription: %s: %w", sub, err)
			}
		}
	}

	if magnetLinks, ok := r.MultipartForm.Value["magnets"]; ok {
		for _, magnetLink := range magnetLinks {
			for _, magnetLink := range strings.Split(magnetLink, "\n") {
				if magnetLink == "" {
					continue
				}

				newTorrent, err := client.TorrentAdd(context.TODO(), transmissionrpc.TorrentAddPayload{
					Filename: &magnetLink,
				})
				if err != nil {
					return err
				}

				newTorrents = append(newTorrents, newTorrent)
			}
		}
	}

	if torrentFileHeaders, ok := r.MultipartForm.File["torrents"]; ok {
		for _, torrentFileHeader := range torrentFileHeaders {
			torrentFile, err := torrentFileHeader.Open()
			if err != nil {
				return err
			}
			defer torrentFile.Close()

			var metaInfo []byte
			if metaInfo, err = io.ReadAll(torrentFile); err != nil {
				return err
			}

			metaInfoEnc := base64.StdEncoding.EncodeToString(metaInfo)

			newTorrent, err := client.TorrentAdd(context.TODO(), transmissionrpc.TorrentAddPayload{
				MetaInfo: &metaInfoEnc,
			})
			if err != nil {
				return err
			}

			newTorrents = append(newTorrents, newTorrent)
		}
	}

	for _, newTorrent := range newTorrents {
		newTorrent := newTorrent

		torrents.Add(Torrent{
			Details: &newTorrent,
			Session: session,
			Push:    push,
		})
	}

	return nil
}

func handle(w http.ResponseWriter, r *http.Request) {
	var session string
	var err error

	w.Header().Add("Content-Type", "text/html")

	if cookie, err := r.Cookie("session"); errors.Is(err, http.ErrNoCookie) {
		session = RandomString(16)

		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   session,
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})
	} else {
		session = cookie.Value
	}

	if r.Method == "POST" {
		if err = handleAddTorrent(r, session); err != nil {
			log.Printf("Error: %s\n", err)
		}
	}

	torrents := torrents.FilterBySession(session)

	tpl.Execute(w, TemplateArgs{
		Torrents:       torrents,
		Error:          err,
		Request:        r,
		VAPIDPublicKey: vapid.PublicKey,
	})
}

func parseArguments() {
	if v := os.Getenv("TRANSMISSION_RPC_HOST"); v != "" {
		rpcHost = v
	}

	if v := os.Getenv("TRANSMISSION_RPC_USER"); v != "" {
		rpcUser = v
	}

	if v := os.Getenv("TRANSMISSION_RPC_PASS"); v != "" {
		rpcPass = v
	}

	if v := os.Getenv("TRANSMISSION_RPC_TLS"); v != "" {
		rpcTLS = v != "0" && v != "false"
	}

	if v := os.Getenv("TRANSMISSION_RPC_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err != nil {
			panic(err)
		} else {
			rpcPort = port
		}
	}

	if v := os.Getenv("LISTEN_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err != nil {
			panic(err)
		} else {
			listenPort = port
		}
	}

	if v := os.Getenv("BASE_URI"); v != "" {
		baseURI = v
	}
}

func main() {
	var err error

	// Setup signal handlers
	captureSignal := make(chan os.Signal, 1)
	signal.Notify(captureSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
	go signalHandler(captureSignal)

	// Parse templates
	if tpl, err = template.ParseFS(tplFiles, "templates/index.html"); err != nil {
		log.Fatalf("Failed to parse templates: %s", err)
	}

	// Parse arguments
	parseArguments()

	// Create transmission RPC client
	if client, err = transmissionrpc.New(rpcHost, rpcUser, rpcPass, &transmissionrpc.AdvancedConfig{
		HTTPS: rpcTLS,
		Port:  uint16(rpcPort),
	}); err != nil {
		log.Fatalf("Failed to create RPC client: %s", err)
	}

	ok, svrVersion, _, err := client.RPCVersion(context.Background())
	if !ok {
		log.Fatalln("Failed to connect to RPC service")
	} else if err != nil {
		log.Fatalf("Failed to get server version: %s", err)
	}

	vapid, err = LoadOrCreateVAPIDKeyPair(vapidFilename)
	if err != nil {
		log.Fatalf("Failed to load or generate VAPID keys: %s", err)
	}

	if err := torrents.Load(torrentsFilename); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Failed to load torrents: %s", err)
	}

	go torrents.Watch()

	log.Printf("Connected to %s running Transmission Daemon v%d\n", rpcHost, svrVersion)
	log.Printf("Listening on http://localhost:%d", listenPort)
	log.Printf("VAPID public key: %s", vapid.PublicKey)
	log.Printf("Torrents loaded: %d", torrents.Length())

	subAssets, _ := fs.Sub(assetFiles, "assets")
	assetHandler := http.FileServer(http.FS(subAssets))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			assetHandler.ServeHTTP(w, r)
		} else {
			handle(w, r)
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), http.DefaultServeMux); err != nil {
		log.Fatalf("Failed to listen: %s", err)
	}
}

func signalHandler(signals chan os.Signal) {
	for range signals {
		log.Printf("Shutting down")

		if err := torrents.Save(torrentsFilename); err != nil {
			log.Fatalf("Failed to save torrents: %s", err)
		}

		os.Exit(0)
	}
}
