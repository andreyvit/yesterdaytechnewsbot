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
	TagRenames      map[string]string
	TrimTagPrefixes []string
	StickyLinks     []string
	Categories      []*Category
}

type Post struct {
	URL         string
	Title       string
	TitleIsLink bool
	Time        time.Time
	Category    *Category
	Tags        []string
	Description []*Region
	Links       map[string]string
	StickyLinks []Link
}

type Link struct {
	Key string
	URL string
}

const (
	LinkNameHN = "HN"
)

var (
	hckrnewsRe = regexp.MustCompile(`^https://news.ycombinator.com/item\?id=\d+$`)
	linkRe     = regexp.MustCompile(`^([A-Za-z0-9_-]+): (https?://.*)$`)
)

func parsePost(pp *pinboard.Post, opt ContentOptions) (*Post, error) {
	post := &Post{
		URL:   pp.URL,
		Title: pp.Title,
		Time:  pp.Time,
		Links: make(map[string]string),
	}

	desc, links := parseTrailingLinks(pp.Description)
	post.Links = links

	post.Description = ParseExplicitLinks(strings.TrimSpace(desc), links)

	for _, key := range opt.StickyLinks {
		if url, ok := links[key]; ok {
			post.StickyLinks = append(post.StickyLinks, Link{
				Key: key,
				URL: url,
			})
		}
	}

	tags := []string(pp.Tags)
	post.Category = DetermineCategoryByTags(opt.Categories, tags)
	if post.Category != nil {
		tags = removeTags(tags, post.Category.Tags)
	}
	tags = renameTags(tags, opt.TagRenames)
	post.Tags = parseTags(tags, opt)

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

func parseTags(tags []string, opt ContentOptions) []string {
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
		buf.WriteString(telegramBold(telegramLink(telegram.Escape(p.Title), p.URL)))
		buf.WriteByte('\n')
	} else {
		if p.Title != "" {
			buf.WriteString(telegramBold(telegram.Escape(p.Title)))
			buf.WriteByte('\n')
		}
		buf.WriteString(telegramLink(telegram.Escape(prettifyURL(p.URL)), p.URL))
		buf.WriteByte('\n')
	}
	if d := strings.TrimSpace(buildDescriptionMarkdown(p)); d != "" {
		buf.WriteByte('\n')
		buf.WriteString(d)
		buf.WriteByte('\n')
	}
	return buf.String()
}

func telegramBold(text string) string {
	return fmt.Sprintf("*%s*", text)
}

func telegramLink(title, url string) string {
	return fmt.Sprintf("[%s](%s)", title, telegram.Escape(url))
}

func buildDescriptionMarkdown(p *Post) string {
	linksUsed := make(map[string]bool)

	var descBuilder strings.Builder
	for _, r := range p.Description {
		if r.PrimaryOccurrance && r.LinkKey != "" {
			linksUsed[r.LinkKey] = true
		}
		if r.PrimaryOccurrance && r.LinkValue != "" {
			descBuilder.WriteString(telegramLink(telegram.EscapeExceptFormatting(r.Text), r.LinkValue))
		} else {
			descBuilder.WriteString(telegram.EscapeExceptFormatting(r.Text))
		}
	}
	desc := descBuilder.String()

	// TODO: format explicit links

	var tags []string
	if p.Category != nil {
		tags = append(tags, p.Category.PreferredTag())
	}
	tags = append(tags, p.Tags...)

	var trailers []string
	for _, link := range p.StickyLinks {
		trailers = append(trailers, telegramLink(telegram.Escape(strings.ReplaceAll(link.Key, "_", " ")), link.URL))
	}
	if len(tags) > 0 {
		trailers = append(trailers, telegram.Escape(buildTags(tags)))
	}

	isMultiParagraph := strings.Contains(desc, "\n\n")
	if fragment := strings.Join(trailers, " · "); fragment != "" {
		if isMultiParagraph || isSpecialLine(lastLine(splitLines(desc))) {
			desc = desc + "\n\n" + fragment
		} else if desc != "" {
			desc = desc + "\n" + fragment
		} else {
			desc = fragment
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
		buf.WriteString(strings.ReplaceAll(s, "-", "_"))
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

func makeTagSet(tags []string) map[string]bool {
	m := make(map[string]bool, len(tags))
	for _, tag := range tags {
		m[tag] = true
	}
	return m
}

func removeTags(sourceTags []string, removedTags []string) []string {
	removedTagSet := makeTagSet(removedTags)
	result := make([]string, 0, len(sourceTags))
	for _, tag := range sourceTags {
		if !removedTagSet[tag] {
			result = append(result, tag)
		}
	}
	return result
}

func renameTags(sourceTags []string, renames map[string]string) []string {
	result := make([]string, 0, len(sourceTags))
	for _, tag := range sourceTags {
		if s, ok := renames[tag]; ok {
			if s == "" {
				continue
			} else {
				tag = s
			}
		}
		result = append(result, tag)
	}
	return result
}
