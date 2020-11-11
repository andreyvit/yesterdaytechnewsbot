package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/pinboard"
	"github.com/andreyvit/yesterdaytechnewsbot/internal/telegram"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	conf := Configuration{
		Pinboard: pinboard.Options{
			Credentials: pinboard.Credentials{
				Username: needEnvString("PINBOARD_USER"),
				Password: needEnvString("PINBOARD_PASSWORD"),
			},
		},
		Telegram: telegram.Options{
			Credentials: telegram.Credentials{
				BotToken:    needEnvString("TELEGRAM_BOT_TOKEN"),
				ChannelName: needEnvString("TELEGRAM_CHANNEL_NAME"),
			},
			DryMode: needEnvBool("TELEGRAM_DRY_RUN"),
		},
		Content: ContentOptions{
			MarkerTag:       "ytn",
			SkipTags:        []string{},
			TrimTagPrefixes: []string{"ytn-"},
		},
		StateFile: needEnvString("BOT_STATE_PATH"),
	}

	if s := needEnvString("PINBOARD_MOCK_DATA"); s != "" {
		data, err := ioutil.ReadFile(s)
		if err != nil {
			log.Fatalf("** Cannot read file specified by PINBOARD_MOCK_DATA: %v", err)
		}
		conf.Pinboard.MockData = data
	}

	err := Run(conf)
	if err != nil {
		log.Fatalf("** %v", err)
	}
}

func needEnvString(key string) string {
	s := os.Getenv(key)
	if s == "" {
		log.Fatalf("** Missing value for environment variable %s", key)
	}
	return s
}

func needEnvBool(key string) bool {
	s := os.Getenv(key)
	if s == "" {
		log.Fatalf("** Missing value for environment variable %s", key)
		panic("unreachable")
	}
	switch strings.ToLower(s) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		log.Fatalf("** Invalid value of environment variable %s, expected a boolean: %q", key, s)
		panic("unreachable")
	}
}
