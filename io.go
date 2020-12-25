package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/eiannone/keyboard"
)

type IO struct {
}

func NewIO() *IO {
	return &IO{}
}

func (io *IO) Prompt(prompt string, defaultChoice, cancelChoice rune, choices ...string) rune {
	defaultIndex := -1
	cancelIndex := -1
	var validRunes []rune
	for i, choice := range choices {
		upperIdx := strings.IndexFunc(choice, unicode.IsUpper)
		if upperIdx < 0 {
			continue
		}
		r, _ := utf8.DecodeRuneInString(choice[upperIdx:])
		validRunes = append(validRunes, r)
		if r == defaultChoice {
			defaultIndex = i
		}
		if r == cancelChoice {
			cancelIndex = i
		}
	}

	prompt = fmt.Sprintf("%s [%s] ", prompt, strings.Join(choices, " / "))

	w := os.Stderr
	for {
		fmt.Fprint(w, prompt)

		r, key, err := keyboard.GetSingleKey()
		if err != nil {
			panic(err)
		}

		if r == '\n' {
			if defaultChoice != 0 {
				if defaultIndex >= 0 {
					fmt.Fprintf(w, "%s\n", choices[defaultIndex])
				}
				return defaultChoice
			} else {
				continue
			}
		} else if key == keyboard.KeyEsc {
			if cancelChoice != 0 {
				if cancelIndex >= 0 {
					fmt.Fprintf(w, "%s\n", choices[cancelIndex])
				}
				return cancelChoice
			} else {
				continue
			}
		}

		r = unicode.ToUpper(r)
		for i, vr := range validRunes {
			if r == vr {
				fmt.Fprintf(w, "%s\n", choices[i])
				return r
			}
		}

		fmt.Fprintln(w)
	}
}
