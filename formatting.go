package main

import (
	"regexp"
	"strings"
)

type Region struct {
	Text              string
	LinkKey           string
	LinkValue         string
	PrimaryOccurrance bool
	Explicit          bool
}

var squareBracketedLinkRe = regexp.MustCompile(`\[([^\]]+)\]`)

func ParseExplicitLinks(text string, links map[string]string) []*Region {
	linkRegexps := make([]*regexp.Regexp, 0, len(links))
	linkKeys := make([]string, 0, len(links))
	for key := range links {
		linkRegexps = append(linkRegexps, BuildLinkRegexp(key))
		linkKeys = append(linkKeys, key)
	}
	squareBracketedLinkIndex := len(linkRegexps)
	allRegexps := append(linkRegexps, squareBracketedLinkRe)

	var regions []*Region
	start := 0
	end := len(text)
	for start < end {
		rem := text[start:]
		i, match := findMultipleRegexpLeftmostStringSubmatchIndex(rem, allRegexps)
		if match == nil {
			match = []int{len(rem), len(rem)} // make prefix handling easier
		}

		matchStart, matchEnd := match[0], match[1]
		if matchStart > 0 {
			regions = append(regions, &Region{
				Text: rem[:matchStart],
			})
		}

		if i == squareBracketedLinkIndex {
			r := &Region{
				Text:     rem[match[2]:match[3]],
				Explicit: true,
			}
			regions = append(regions, r)

			keyIndex, _ := findMultipleRegexpLeftmostStringSubmatchIndex(rem, linkRegexps)
			if keyIndex >= 0 {
				r.LinkKey = linkKeys[keyIndex]
				r.LinkValue = links[linkKeys[keyIndex]]
			}

		} else if i >= 0 {
			regions = append(regions, &Region{
				Text:      rem[matchStart:matchEnd],
				LinkKey:   linkKeys[i],
				LinkValue: links[linkKeys[i]],
			})
		}

		start += matchEnd
	}

	primaryFound := make(map[string]bool)
	// log.Printf("=== %s ===", text)
	// for i, r := range regions {
	// 	log.Printf("region %d: %#v", i, *r)
	// }

	for _, r := range regions {
		if r.LinkKey != "" && !primaryFound[r.LinkKey] && r.Explicit {
			r.PrimaryOccurrance = true
			primaryFound[r.LinkKey] = true
		}
	}
	for _, r := range regions {
		if r.LinkKey != "" && !primaryFound[r.LinkKey] {
			r.PrimaryOccurrance = true
			primaryFound[r.LinkKey] = true
		}
	}
	// log.Println("---")
	// for i, r := range regions {
	// 	log.Printf("region %d: %#v", i, *r)
	// }

	return regions
}

func findMultipleRegexpLeftmostStringSubmatchIndex(s string, res []*regexp.Regexp) (regexpIndex int, match []int) {
	regexpIndex = -1
	for i, re := range res {
		if m := re.FindStringSubmatchIndex(s); m != nil {
			if match == nil || m[0] < match[0] {
				regexpIndex = i
				match = m
			}
		}
	}
	return
}

func BuildLinkRegexp(link string) *regexp.Regexp {
	regions := strings.Split(link, "_")
	for i, region := range regions {
		regions[i] = regexp.QuoteMeta(region)
	}
	return regexp.MustCompile(`(?i)\b` + strings.Join(regions, `[\s._]{0,2}`) + `\b`)
}
