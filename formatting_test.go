package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseExplicitLinks(t *testing.T) {
	tests := []struct {
		Input    string
		Links    string
		Expected string
	}{
		{"Hello HN world", "", "Hello HN world"},
		{"Hello HN world", "HN", "Hello |HN<HN>| world"},
		{"Hello HN and HN world", "HN", "Hello |HN<HN>| and |HN{HN}| world"},
		{"Hello HN world", "foo HN world bar", "Hello |HN<HN>| |world<world>"},
		{"Hello HN world!", "foo HN world bar", "Hello |HN<HN>| |world<world>|!"},
		{"Hello HN World!", "foo HN_world bar", "Hello |HN World<HN_world>|!"},
		{"Hello [virtual HN world]!", "foo hn bar", "Hello |virtual HN world<hn>|!"},
		{"[Hello HN world]!", "foo HN bar", "Hello HN world<HN>|!"},
		{"[Hello foo world]", "foo hn bar", "Hello foo world<foo>"},
		{"Hello, [crazy foobar world]!", "foo hn bar", "Hello, |crazy foobar world|!"},
		{"Hello, [crazy foobar HN world]!", "foo hn bar", "Hello, |crazy foobar HN world<hn>|!"},
		{"Hello, [crazy bar foobar world]!", "foo hn bar", "Hello, |crazy bar foobar world<bar>|!"},
		{"Hello, [crazy foo foobar world]!", "foo hn bar", "Hello, |crazy foo foobar world<foo>|!"},
		{"Hello, [crazy foobar bar world]!", "foo hn bar", "Hello, |crazy foobar bar world<bar>|!"},
		{"Hello, [crazy foobar foo world]!", "foo hn bar", "Hello, |crazy foobar foo world<foo>|!"},
		// {"Hello, [unrelated](somewhere)!", "foo hn bar", "Hello, |unrelated<:somewhere>|!"},
	}
	for _, test := range tests {
		links := make(map[string]string)
		for _, key := range strings.Fields(test.Links) {
			links[key] = fmt.Sprintf("https://example.com/%s/", key)
		}
		regions := ParseExplicitLinks(test.Input, links)
		actual := DescribeRegions(regions)

		if actual != test.Expected {
			t.Errorf("ParseExplicitLinks(%q, %q) = %q, wanted %q", test.Input, test.Links, actual, test.Expected)
		}
	}
}

func DescribeRegions(regions []*Region) string {
	var buf strings.Builder
	for i, r := range regions {
		if i > 0 {
			buf.WriteByte('|')
		}
		buf.WriteString(r.Text)
		if r.LinkKey != "" {
			if r.PrimaryOccurrance {
				buf.WriteByte('<')
			} else {
				buf.WriteByte('{')
			}
			buf.WriteString(r.LinkKey)
			if r.PrimaryOccurrance {
				buf.WriteByte('>')
			} else {
				buf.WriteByte('}')
			}
		} else if r.LinkValue != "" {
			buf.WriteByte('<')
			buf.WriteByte(':')
			buf.WriteString(r.LinkValue)
			buf.WriteByte('>')
		}
	}
	return buf.String()
}
