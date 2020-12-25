package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	httpsimp "github.com/andreyvit/httpsimplified/v2"
	"github.com/andreyvit/yesterdaytechnewsbot/internal/curlstr"
)

type Credentials struct {
	BotToken    string
	ChannelName string
}

type Options struct {
	Credentials
	DryMode bool
}

type Message struct {
	MarkdownText     string
	EnableWebPreview bool
}

const (
	baseURL = "https://api.telegram.org"
)

var tgBool = map[bool]string{false: "0", true: "1"}

func PostText(msg *Message, opt Options) error {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	r := httpsimp.MakeGet(baseURL, fmt.Sprintf("/bot%s/sendMessage", opt.BotToken), url.Values{
		"chat_id":                  []string{"@" + opt.ChannelName},
		"text":                     []string{msg.MarkdownText},
		"parse_mode":               []string{"MarkdownV2"},
		"disable_web_page_preview": []string{tgBool[!msg.EnableWebPreview]},
	}, nil)

	log.Printf("[telegram] $ %s", curlstr.CurlString(r))
	if opt.DryMode {
		log.Printf("[telegram] dry mode for message:\n%s", indent(msg.MarkdownText))
	} else {
		log.Printf("[telegram] sending message:\n%s", indent(msg.MarkdownText))

		var resp sendMessageResponse
		err := httpsimp.Do(r, client, httpsimp.JSON(&resp))
		if err != nil {
			return err
		}

		if !resp.OK {
			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				panic(err)
			}
			log.Printf("WARNING: telegram post failed: %s", data)
			return fmt.Errorf("telegram post failed")
		}
	}

	return nil
}

type sendMessageResponse struct {
	OK          bool        `json:"ok"`
	Result      interface{} `json:"result"`
	ErrorCode   int         `json:"error_code"`
	Description string      `json:"description"`
}

func Escape(s string) string {
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "`", "\\`")
	return EscapeExceptFormatting(s)
}

func EscapeExceptFormatting(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "-", "\\-")
	s = strings.ReplaceAll(s, "+", "\\+")
	s = strings.ReplaceAll(s, ".", "\\.")
	s = strings.ReplaceAll(s, "#", "\\#")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "~", "\\~")
	s = strings.ReplaceAll(s, ">", "\\>")
	s = strings.ReplaceAll(s, "=", "\\=")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	return s
}

const indentStep = "    "

func indent(s string) string {
	if s == "" {
		return ""
	}
	return indentStep + strings.ReplaceAll(s, "\n", "\n"+indentStep)
}
