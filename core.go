package main

import (
	"fmt"
	"log"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/pinboard"
	"github.com/andreyvit/yesterdaytechnewsbot/internal/telegram"
)

type Configuration struct {
	Pinboard  pinboard.Options
	Telegram  telegram.Options
	Content   ContentOptions
	StateFile string
}

func Run(conf Configuration) error {
	posts, err := pinboard.LoadRecent(pinboard.RecentRequest{}, conf.Pinboard)
	if err != nil {
		return err
	}

	for _, post := range posts {
		if !post.Tags.Contains("ytn") {
			log.Println()
			log.Printf("IGNORING: %s", post.TitleOrURL())
			continue
		}
		err := handle(post, conf)
		if err != nil {
			return fmt.Errorf("%v [while handling: %s]", err, post.TitleOrURL())
		}
	}

	return nil
}

func handle(pp *pinboard.Post, conf Configuration) error {
	log.Println()
	log.Printf("HANDLING POST:\n%v\n", pp)

	post, err := parsePost(pp, conf.Content)
	if err != nil {
		return err
	}

	msg := &telegram.Message{
		// MarkdownText: fmt.Sprintf("Hello **world!**\nOn %s", time.Now().Format(time.RFC3339)),
		MarkdownText: buildTelegramMarkdown(post),
	}
	err = telegram.PostText(msg, conf.Telegram)
	if err != nil {
		return err
	}

	return nil
}
