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

const posts = `<?xml version="1.0" encoding="UTF-8" ?>
<posts user="andreyvit" dt="2020-11-04T12:31:49Z">
    <post href="https://ottverse.com/eme-cenc-cdm-aes-keys-drm-digital-rights-management/" time="2020-11-04T12:31:49Z" description="EME, CDM, AES, CENC, and Keys - The Essential Building Blocks of DRM - OTTVerse" extended="" tag="drm" hash="fa05e017d1c0966d88bc8bc1c7f61912"    />
    <post href="https://vlfig.me/posts/microservices" time="2020-11-04T12:26:24Z" description="Microservices — architecture nihilism in minimalism's clothes - Blog by Vasco Figueira" extended="" tag="architecture soa" hash="35d4250cb7fbdde8094d077e8148fe37"    />
    <post href="https://www.quantamagazine.org/a-new-map-of-the-standard-model-of-particle-physics-20201022/" time="2020-11-01T21:22:00Z" description="A New Map of the Standard Model of Particle Physics | Quanta Magazine" extended="" tag="physics" hash="dcf6350f20c5f4f58d45b2f5cbf69ed5"    />
    <post href="https://habr.com/ru/post/410883/" time="2020-10-25T09:49:28Z" description="Дюжина советов – как научить ребенка шахматам. И не только / Хабр" extended="" tag="russian chess kids" hash="f84b13c6948dc126586972d611cd6076"    />
    <post href="https://www.sciencedirect.com/science/article/pii/S0924977X20309111" time="2020-10-23T20:18:06Z" description="Mood and cognition after administration of low LSD doses in healthy volunteers: A placebo controlled dose-effect finding study - ScienceDirect" extended="https://news.ycombinator.com/item?id=24857679" tag="health drugs" hash="dc984cf947d8c0fe60b0c662fed555b3"    />
    <post href="https://catchjs.com/Blog/Churn" time="2020-10-21T11:30:13Z" description="You're all calculating churn rates wrong | CatchJS" extended="https://news.ycombinator.com/item?id=24831637" tag="business metrics statistics analytics" hash="2b3c9e8063c5247276fd3d41fecbf394"    />
    <post href="http://blog.rlmflores.me/2020/10/14/what_is_expected_of_an_engineering_manager/" time="2020-10-17T12:07:10Z" description="What is expected of a Engineering Manager? · Rodrigo Flores's Corner" extended="https://news.ycombinator.com/item?id=24787002" tag="management" hash="5022866c8b290a63984ef42a458adc4a"    />
    <post href="https://www.comet.ml/site/why-software-engineering-processes-and-tools-dont-work-for-machine-learning/" time="2020-10-16T09:13:44Z" description="Why software engineering processes and tools don’t work for machine learning – Comet" extended="" tag="ai management" hash="b78341a876519c858ec7936363461f1d"    />
    <post href="https://www.optaplanner.org/" time="2020-10-14T15:02:29Z" description="OptaPlanner - Constraint satisfaction solver (Java™, Open Source)" extended="" tag="" hash="2e1f3249685b3acfe41ca1055f4d5688"    />
    <post href="https://blog.eyas.sh/2020/10/unity-for-engineers-pt1-basic-concepts/" time="2020-10-13T10:37:49Z" description="Basic Concepts in Unity for Software Engineers | Eyas's Blog" extended="" tag="" hash="f29bc518160e272bf6d1e327bd21d0d5"   toread="yes" />
    <post href="https://artvee.com/" time="2020-10-12T13:03:11Z" description="Artvee" extended="" tag="stock artwork" hash="64397732c03cf1b96c3500306241fe2d"    />
    <post href="https://cchound.com/" time="2020-10-12T12:46:28Z" description="cchound.com | free music for content creators" extended="" tag="stock music" hash="5e10f3b22c33339362de456c8d191f0b"    />
    <post href="http://staytus.co/" time="2020-10-12T10:36:07Z" description="The Open Source Status Site - Staytus" extended="" tag="monitoring deployment" hash="06a25b03fbb4c02c41466e70507567a1"    />
    <post href="http://www.righto.com/2011/07/cells-are-very-fast-and-crowded-places.html" time="2020-10-06T13:39:27Z" description="Cells are very fast and crowded places" extended="" tag="biology" hash="9a002eab1ec30069a9c8ad62e6f69f11"    />
    <post href="https://github.com/kdeldycke/awesome-falsehood" time="2020-09-09T09:04:18Z" description="kdeldycke/awesome-falsehood: 😱 Falsehoods Programmers Believe in" extended="" tag="programming" hash="c90f0487a9b034e1083f0c0f4cac3ec7"    />
</posts>`

func LoadRecent(req RecentRequest, opt Options) ([]*Post, error) {
	var resp postsResponse
	if false {
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

		log.Printf("> %s", curlstr.CurlString(r))

		err := httpsimp.Do(r, client, XML(&resp))
		if err != nil {
			return nil, err
		}

	} else {
		err := xml.Unmarshal([]byte(posts), &resp)
		if err != nil {
			return nil, err
		}
		log.Printf("resp = %#v", resp)
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
