package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/hekmon/transmissionrpc/v2"
)

type Torrent struct {
	Hash    string                   `json:"hash"`
	Session string                   `json:"session"`
	Push    *webpush.Subscription    `json:"push,omitempty"`
	Details *transmissionrpc.Torrent `json:"-"`
}

func (t *Torrent) SendNotification() error {
	if t.Push == nil {
		return nil
	}

	notif := Notification{
		Title: "Finished downloading Torrent",
		Body:  fmt.Sprintf("Name: %s", *t.Details.Name),
		Icon:  fmt.Sprintf("%s/logo.svg", baseURI),
	}

	msg, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification(msg, t.Push, &webpush.Options{
		Subscriber:      "example@example.com",
		TTL:             60 * 60 * 24,
		VAPIDPublicKey:  vapid.PublicKey,
		VAPIDPrivateKey: vapid.PrivateKey,
	})
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

type TorrentStore struct {
	torrents map[string]Torrent
	mu       sync.Mutex
}

func NewTorrentStore() TorrentStore {
	return TorrentStore{
		torrents: map[string]Torrent{},
	}
}

func (s *TorrentStore) Load(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	torrents := []Torrent{}

	if err := json.Unmarshal(data, &torrents); err != nil {
		return err
	}

	s.mu.Lock()
	for _, torrent := range torrents {
		torrent := torrent
		s.torrents[torrent.Hash] = torrent
	}
	s.mu.Unlock()

	return s.Sync()
}

func (s *TorrentStore) Save(fn string) error {
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	l := []Torrent{}

	s.mu.Lock()
	for _, torrent := range s.torrents {
		torrent := torrent
		torrent.Hash = *torrent.Details.HashString
		l = append(l, torrent)
	}
	s.mu.Unlock()

	return json.NewEncoder(f).Encode(l)
}

func (s *TorrentStore) FilterBySession(session string) []Torrent {
	ss := []Torrent{}

	s.mu.Lock()
	for _, torrent := range s.torrents {
		torrent := torrent

		if torrent.Session == session {
			ss = append(ss, torrent)
		}
	}
	s.mu.Unlock()

	return ss
}

func (s *TorrentStore) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hashes := []string{}
	for hash := range s.torrents {
		hashes = append(hashes, hash)
	}

	newTorrents, err := client.TorrentGetAllForHashes(context.Background(), hashes)
	if err != nil {
		return err
	}

	newTorrentsMap := map[string]Torrent{}

	for _, newTorrent := range newTorrents {
		newTorrent := newTorrent

		oldTorrent, ok := s.torrents[*newTorrent.HashString]
		if !ok {
			continue
		}

		oldTorrent.Details = &newTorrent
		newTorrentsMap[*oldTorrent.Details.HashString] = oldTorrent
	}

	s.torrents = newTorrentsMap

	return nil
}

func (s *TorrentStore) Watch() {
	t := time.NewTicker(1 * time.Second)
	for range t.C {
		if err := s.Sync(); err != nil {
			log.Printf("Failed to sync torrents: %s", err)
			continue
		}

		s.Notify()
	}
}

func (s *TorrentStore) Notify() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, torrent := range s.torrents {
		torrent := torrent

		if torrent.Details.PercentDone != nil && *torrent.Details.PercentDone >= 1.0 {
			log.Printf("Finished downloading %s", *torrent.Details.Name)

			if err := torrent.SendNotification(); err != nil {
				log.Printf("Failed to send notification: %s\n", err)
				continue
			}

			delete(s.torrents, torrent.Hash)
		}
	}
}

func (s *TorrentStore) Add(torrent Torrent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if torrent.Details.HashString != nil {
		torrent.Hash = *torrent.Details.HashString
	}

	log.Printf("Added torrent %s", *torrent.Details.Name)

	s.torrents[*torrent.Details.HashString] = torrent
}

func (s *TorrentStore) Length() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.torrents)
}
