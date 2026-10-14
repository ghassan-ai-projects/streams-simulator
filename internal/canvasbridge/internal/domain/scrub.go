package domain

import "strings"

func Scrub(text, root string) string {
	if root == "" {
		return text
	}
	return strings.ReplaceAll(strings.ReplaceAll(text, root+"/", ""), root, ".")
}
