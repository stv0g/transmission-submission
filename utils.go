package main

import "regexp"

func decodeLabels(labels []string) map[string]string {
	decLabels := map[string]string{}

	re := regexp.MustCompile(`^x-([a-zA-Z0-9]+)=([a-zA-Z0-9]+)$`)

	for _, label := range labels {
		if m := re.FindStringSubmatch(label); m != nil {
			decLabels[m[1]] = m[2]
		}
	}

	return decLabels
}
