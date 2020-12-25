package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/google/renameio"
)

const (
	ChannelTelegram = "tg"
)

type State struct {
	PublishedArticles map[string]*ArticleState `json:"published_articles"`
}

func (state *State) LookupArticle(url string) *ArticleState {
	h := HashOfURL(url)
	as := state.PublishedArticles[h]
	if as == nil {
		as = &ArticleState{
			URL: url,
		}
		state.PublishedArticles[h] = as
	}
	if as.Channels == nil {
		as.Channels = make(map[string]*ArticleChannelState)
	}
	return as
}

type ArticleState struct {
	URL      string                          `json:"url"`
	Skip     bool                            `json:"skip,omitempty"`
	Channels map[string]*ArticleChannelState `json:"ch"`
}

type ArticleChannelState struct {
	PublishTime time.Time `json:"t"`
}

func ReadState(fn string) (*State, error) {
	raw, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("load state: zero-length state file at %s", fn)
	}

	state := new(State)
	err = json.Unmarshal(raw, state)
	if err != nil {
		return nil, err
	}

	if state.PublishedArticles == nil {
		state.PublishedArticles = make(map[string]*ArticleState)
	}

	return state, nil
}

func WriteState(fn string, state *State) error {
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		panic(err)
	}

	err = renameio.WriteFile(fn, raw, 0644)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}
