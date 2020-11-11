package pinboard

import (
	"encoding/xml"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	httpsimp "github.com/andreyvit/httpsimplified/v2"

	"github.com/andreyvit/yesterdaytechnewsbot/internal/curlstr"
)

type Credentials struct {
	Username string
	Password string
}

func (cred Credentials) authHeaders() http.Header {
	return http.Header{
		httpsimp.AuthorizationHeader: []string{httpsimp.BasicAuthValue(cred.Username, cred.Password)},
	}
}

type Options struct {
	Credentials
	MockData []byte
}

type TagList []string

func (tl TagList) String() string {
	var buf strings.Builder
	for i, s := range tl {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteByte('#')
		buf.WriteString(s)
	}
	return buf.String()
}

func (tl TagList) Index(tag string) int {
	for i, t := range tl {
		if t == tag {
			return i
		}
	}
	return -1
}

func (tl TagList) Contains(tag string) bool {
	return tl.Index(tag) != -1
}

type Post struct {
	URL         string
	Title       string
	Time        time.Time
	Tags        TagList
	Description string
}

func (p *Post) TitleOrURL() string {
	if p.Title != "" {
		return p.Title
	}
	return p.URL
}

func (p *Post) String() string {
	var buf strings.Builder
	if p.Title != "" {
		buf.WriteString(p.Title)
		buf.WriteByte('\n')
	}
	buf.WriteString(p.URL)
	buf.WriteByte('\n')
	if s := p.Tags.String(); s != "" {
		buf.WriteString(s)
		buf.WriteByte('\n')
	}
	if p.Description != "" {
		buf.WriteByte('\n')
		buf.WriteString(strings.TrimSpace(p.Description))
		buf.WriteByte('\n')
	}
	return buf.String()
}

const (
	baseURL = "https://api.pinboard.in/v1"
)

type RecentRequest struct {
	Tag   string
	Limit int
}

func LoadRecent(req RecentRequest, opt Options) ([]*Post, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	params := make(url.Values)
	if req.Tag != "" {
		params.Set("tag", req.Tag)
	}
	if req.Limit != 0 {
		params.Set("count", strconv.Itoa(req.Limit))
	}

	r := httpsimp.MakeGet(baseURL, "/posts/recent", params, opt.authHeaders())

	log.Printf("[pinboard] %s", curlstr.CurlString(r))

	var resp postsResponse
	if opt.MockData != nil {
		err := xml.Unmarshal(opt.MockData, &resp)
		if err != nil {
			return nil, err
		}
	} else {
		err := httpsimp.Do(r, client, XML(&resp))
		if err != nil {
			return nil, err
		}
	}
	return mapPosts(resp.Posts), nil
}

func mapPosts(pps []postPayload) []*Post {
	var posts []*Post
	for _, pp := range pps {
		posts = append(posts, mapPost(pp))
	}
	return posts
}

func mapPost(pp postPayload) *Post {
	return &Post{
		URL:         pp.URL,
		Title:       pp.Title,
		Time:        pp.Time,
		Tags:        mapTags(pp.Tags),
		Description: pp.Description,
	}
}

func mapTags(source string) TagList {
	var l TagList
	for _, s := range strings.Fields(source) {
		if s == "" {
			continue
		}
		l = append(l, s)
	}
	return l
}

type postsResponse struct {
	Posts []postPayload `xml:"post"`
}

type postPayload struct {
	URL         string    `xml:"href,attr"`
	Title       string    `xml:"description,attr"`
	Time        time.Time `xml:"time,attr"`
	Tags        string    `xml:"tag,attr"`
	Description string    `xml:"extended,attr"`
}

/*
JSON is a Parser function that verifies the response status code and content
type (which must be ContentTypeJSON) and unmarshals the body into the
result variable (which can be anything that you'd pass to json.Unmarshal).

Pass the result of this function into Do or Parse to handle a response.
*/
func XML(result interface{}, mopt ...httpsimp.ParseOption) httpsimp.Parser {
	if result == nil {
		var body interface{}
		result = &body
	}
	return httpsimp.MakeParser("text/xml", mopt, func(resp *http.Response) (interface{}, error) {
		defer resp.Body.Close()
		err := xml.NewDecoder(resp.Body).Decode(result)
		body := reflect.ValueOf(result).Elem().Interface()
		return body, err
	})
}
