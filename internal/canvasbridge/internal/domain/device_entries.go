package domain

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MaxListedFileBytes = 256 << 10
	maxListedTitle     = 200
)

var (
	kindName      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	componentName = regexp.MustCompile(`^[a-z][a-z0-9-]{1,31}/[a-z][a-z0-9-]{1,31}$`)
)

type KindEntry struct {
	Path  string `json:"path"`
	Kind  string `json:"kind,omitempty"`
	Title string `json:"title,omitempty"`
}

type ComponentEntry struct {
	Path  string    `json:"path"`
	ID    string    `json:"id,omitempty"`
	Title string    `json:"title,omitempty"`
	Class FileClass `json:"class"`
}

func KindEntryOf(rel string, content []byte) KindEntry {
	var head struct {
		Kind  string `json:"kind"`
		Title string `json:"title"`
	}
	if json.Unmarshal(content, &head) != nil {
		return KindEntry{Path: rel}
	}
	return KindEntry{Path: rel, Kind: listedName(head.Kind, kindName), Title: listedTitle(head.Title)}
}

func ComponentEntryOf(rel string, class FileClass, content []byte) ComponentEntry {
	var head struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if json.Unmarshal(content, &head) != nil {
		return ComponentEntry{Path: rel, Class: class}
	}
	return ComponentEntry{Path: rel, ID: listedName(head.ID, componentName), Title: listedTitle(head.Title), Class: class}
}

func listedName(name string, pattern *regexp.Regexp) string {
	if pattern.MatchString(name) {
		return name
	}
	return ""
}

func listedTitle(title string) string {
	if utf8.RuneCountInString(title) > maxListedTitle || strings.ContainsFunc(title, unlistable) {
		return ""
	}
	return title
}

func unlistable(r rune) bool {
	return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)
}
