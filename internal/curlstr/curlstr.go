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

	for k, vv := range r.URL.Query() {
		for _, v := range vv {
			args = append(args, "-d", fmt.Sprintf("'%s=%s'", k, v))
		}
	}
	r.URL.RawQuery = ""

	for k, vv := range r.Header {
		for _, v := range vv {
			args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
		}
	}

	args = append(args, r.URL.String())

	for i, s := range args {
		args[i] = quote(s)
	}
	return strings.Join(args, " ")
}

func quote(v string) string {
	if strings.ContainsAny(v, "'\" $&|;{}") {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "'", "\\'") + "'"
	} else {
		return v
	}
}
