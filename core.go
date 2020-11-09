package main

import (
	"log"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/pinboard"
)

type Configuration struct {
	Pinboard  pinboard.Options
	StateFile string
}

func Run(conf Configuration) error {
	posts, err := pinboard.LoadRecent(pinboard.RecentRequest{}, conf.Pinboard)
	if err != nil {
		return err
	}

	log.Printf("Loaded %d posts:", len(posts))
	for i, p := range posts {
		log.Printf("\nPOST %03d\n%v", i, p)
	}

	return nil
}
