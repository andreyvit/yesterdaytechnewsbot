package main

import (
	"strings"
)

const indentStep = "    "

func indent(s string) string {
	if s == "" {
		return ""
	}
	return indentStep + strings.ReplaceAll(s, "\n", "\n"+indentStep)
}
