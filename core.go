package main

import (
	"fmt"
	"log"
	"time"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/pinboard"
	"github.com/andreyvit/yesterdaytechnewsbot/internal/telegram"
)

type Configuration struct {
	Pinboard     pinboard.Options
	Telegram     telegram.Options
	Content      ContentOptions
	StateFile    string
	RepublishAll bool
}

type Env struct {
	Conf  Configuration
	IO    *IO
	State *State
}

var (
	ErrQuit = fmt.Errorf("quit")
)

func Run(conf Configuration) error {
	env := &Env{
		Conf: conf,
		IO:   NewIO(),
	}

	state, err := ReadState(conf.StateFile)
	if err != nil {
		return err
	}
	env.State = state

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
		err := env.handle(post, conf)
		if err == ErrQuit {
			break
		} else if err != nil {
			return fmt.Errorf("%v [while handling: %s]", err, post.TitleOrURL())
		}
	}

	return nil
}

func (env *Env) handle(pp *pinboard.Post, conf Configuration) error {
	as := env.State.LookupArticle(pp.URL)
	if as.Skip {
		log.Printf("SKIPPED:\n%v\n", pp)
		return nil
	}

	log.Println()
	if as.Channels[ChannelTelegram] != nil {
		if conf.RepublishAll {
			log.Printf("REPUBLISHING:\n%v\n", pp)
		} else {
			log.Printf("ALREADY PUBLISHED:\n%v\n", pp)
			return nil
		}
	} else {
		log.Printf("PUBLISHING:\n%v\n", pp)
	}

	post, err := parsePost(pp, conf.Content)
	if err != nil {
		return err
	}

	if post.Category == nil {
		log.Println()
		log.Printf("NO CATEGORY:\n%v\n", pp)
		return nil
	}

	msg := &telegram.Message{
		// MarkdownText: fmt.Sprintf("Hello **world!**\nOn %s", time.Now().Format(time.RFC3339)),
		MarkdownText: buildTelegramMarkdown(post),
	}

	log.Printf("TELEGRAM MESSAGE:\n%s", indent(msg.MarkdownText))

	switch env.IO.Prompt("Publish to Telegram?", 0, 'L', "Publish", "Later", "Skip permanently", "Quit") {
	case 'P':
		break
	case 'L':
		return nil
	case 'S':
		as.Skip = true
		if err := env.saveState(); err != nil {
			return err
		}
		return nil
	case 'Q':
		return ErrQuit
	default:
		panic("unhandled choice")
	}
	err = telegram.PostText(msg, conf.Telegram)
	if err != nil {
		return err
	}

	as.Channels[ChannelTelegram] = &ArticleChannelState{
		PublishTime: time.Now(),
	}

	if err := env.saveState(); err != nil {
		return err
	}

	return nil
}

func (env *Env) saveState() error {
	return WriteState(env.Conf.StateFile, env.State)
}
