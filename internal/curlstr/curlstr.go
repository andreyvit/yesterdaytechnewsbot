package curlstr

import (
	"fmt"
	"net/http"
	"strings"
)

func CurlString(r *http.Request) string {
	var args []string
	args = append(args, "curl")
	if r.Method != http.MethodGet {
		args = append(args, "-X", r.Method)
	}

	u := *r.URL
	for k, vv := range u.Query() {
		for _, v := range vv {
			args = append(args, "-d", fmt.Sprintf("%s=%s", k, v))
		}
	}
	u.RawQuery = ""

	for k, vv := range r.Header {
		for _, v := range vv {
			args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
		}
	}

	args = append(args, u.String())

	for i, s := range args {
		args[i] = quote(s)
	}
	return strings.Join(args, " ")
}

func quote(v string) string {
	if strings.ContainsAny(v, "'\" $&|;{}\n") {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "'", "\\'") + "'"
	} else {
		return v
	}
}
