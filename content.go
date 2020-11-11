package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/pinboard"
	"github.com/andreyvit/yesterdaytechnewsbot/internal/telegram"
)

type ContentOptions struct {
	MarkerTag       string
	SkipTags        []string
	TrimTagPrefixes []string
}

type Post struct {
	URL         string
	Title       string
	TitleIsLink bool
	Time        time.Time
	Tags        []string
	Description string
	Links       map[string]string
}

const (
	LinkNameHN = "HN"
)

var (
	hckrnewsRe = regexp.MustCompile(`^https://news.ycombinator.com/item\?id=\d+$`)
	linkRe     = regexp.MustCompile(`^(\w+): (https?://.*)$`)
)

func parsePost(pp *pinboard.Post, opt ContentOptions) (*Post, error) {
	post := &Post{
		URL:   pp.URL,
		Title: pp.Title,
		Time:  pp.Time,
		Links: make(map[string]string),
	}

	post.Description, post.Links = parseTrailingLinks(pp.Description)
	post.Tags = parseTags(pp.Tags, opt)

	post.TitleIsLink = strings.HasPrefix(post.URL, "https://news.ycombinator.com/")

	return post, nil
}

func parseTrailingLinks(desc string) (string, map[string]string) {
	links := make(map[string]string)

	if !strings.Contains(desc, "\n\n") {
		// avoid parsing post URL as Hacker News link for posts like Ask HN
		return strings.TrimSpace(desc), links
	}

	lines := strings.Split(strings.TrimSpace(desc), "\n")
	cont := true
	for len(lines) > 0 && cont {
		line := strings.TrimSpace(lines[len(lines)-1])
		if line == "" {
			cont = false
		} else if hckrnewsRe.MatchString(line) {
			links["HN"] = line
		} else if m := linkRe.FindStringSubmatch(line); m != nil {
			links[m[1]] = m[2]
		} else {
			break
		}
		lines = lines[:len(lines)-1]
	}

	return strings.TrimSpace(strings.Join(lines, "\n")), links
}

func parseTags(tags pinboard.TagList, opt ContentOptions) []string {
	skip := make(map[string]bool)
	if opt.MarkerTag != "" {
		skip[opt.MarkerTag] = true
	}
	for _, tag := range opt.SkipTags {
		skip[tag] = true
	}

	seen := make(map[string]bool)

	var result []string
	for _, tag := range tags {
		if skip[tag] {
			continue
		}
		for _, prefix := range opt.TrimTagPrefixes {
			tag = strings.TrimPrefix(tag, prefix)
		}

		if seen[tag] {
			continue
		}
		seen[tag] = true
		result = append(result, tag)
	}

	return result
}

func buildTelegramMarkdown(p *Post) string {
	var buf strings.Builder
	// buf.WriteString(telegram.EscapeReserved(telegram.EscapeForMarkdown(time.Now().Format(time.RFC3339))) + "\n")

	if p.TitleIsLink && p.Title != "" {
		buf.WriteString(telegram.EscapeReserved(telegramLink(p.Title, p.URL)))
		buf.WriteByte('\n')
	} else {
		if p.Title != "" {
			buf.WriteString(telegram.EscapeReserved(telegram.EscapeForMarkdown(p.Title)))
			buf.WriteByte('\n')
		}
		buf.WriteString(telegram.EscapeReserved(telegramLink(prettifyURL(p.URL), p.URL)))
		buf.WriteByte('\n')
	}
	if p.Description != "" {
		buf.WriteByte('\n')
		buf.WriteString(telegram.EscapeReserved(strings.TrimSpace(buildDescriptionMarkdown(p))))
		buf.WriteByte('\n')
	}
	return buf.String()
}

func telegramLink(title, url string) string {
	return fmt.Sprintf("[%s](%s)", telegram.EscapeForMarkdown(title), url)
}

func buildDescriptionMarkdown(p *Post) string {
	desc := strings.TrimSpace(p.Description)

	linksUsed := make(map[string]bool)

	// TODO: format explicit links

	var trailers []string
	for key, url := range p.Links {
		if linksUsed[key] {
			continue
		}
		trailers = append(trailers, fmt.Sprintf("[%s](%s)", telegram.EscapeForMarkdown(key), url))
	}
	if len(p.Tags) > 0 {
		trailers = append(trailers, telegram.EscapeForMarkdown(buildTags(p.Tags)))
	}

	isMultiParagraph := strings.Contains(p.Description, "\n\n")
	if fragment := strings.Join(trailers, " · "); fragment != "" {
		if isMultiParagraph || isSpecialLine(lastLine(splitLines(desc))) {
			desc = desc + "\n\n" + fragment
		} else {
			desc = desc + " " + fragment + "."
		}
	}

	return desc
}

func buildTags(tags []string) string {
	var buf strings.Builder
	for _, s := range tags {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteByte('#')
		buf.WriteString(s)
	}
	return buf.String()
}

func lastLine(lines []string) string {
	n := len(lines)
	if n == 0 {
		return ""
	}
	return lines[n-1]
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func isSpecialLine(line string) bool {
	if len(line) == 0 {
		return true
	}
	c := line[0]
	if c == '*' || c == '-' || c == '>' || c == '#' {
		return true
	}
	return false
}

func prettifyURL(link string) string {
	// if u, err := url.Parse(link); err == nil {
	// 	u
	// }
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "https://")
	return link
}
