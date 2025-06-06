package config

import (
	"bufio"
	"os"
	"strings"
)

type FeedConfig struct {
	URL      string
	Category string
}

type Config struct {
	Feeds []FeedConfig
}

func LoadConfig(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	scanner := bufio.NewScanner(file)
	var mode string
	var feed FeedConfig

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		switch line {
		case "feed:":
			if feed.URL != "" {
				cfg.Feeds = append(cfg.Feeds, feed)
				feed = FeedConfig{}
			}
			mode = "feed"

		default:
			if mode == "feed" {
				if strings.HasPrefix(line, "url:") {
					feed.URL = strings.TrimSpace(strings.TrimPrefix(line, "url:"))
				} else if strings.HasPrefix(line, "category:") {
					feed.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
				}
			}
		}
	}

	if feed.URL != "" {
		cfg.Feeds = append(cfg.Feeds, feed)
	}

	return cfg, scanner.Err()
}
