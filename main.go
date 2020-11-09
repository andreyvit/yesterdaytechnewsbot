package main

import (
	"log"
	"os"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/pinboard"
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
		StateFile: needEnvString("BOT_STATE_PATH"),
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
