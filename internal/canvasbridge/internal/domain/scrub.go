package domain

import (
	"cmp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Scrub(text string, roots []string) string {
	for _, root := range longestFirst(roots) {
		text = scrubRoot(text, root)
	}
	return text
}

func longestFirst(roots []string) []string {
	kept := slices.DeleteFunc(slices.Clone(roots), func(root string) bool { return root == "" })
	slices.SortStableFunc(kept, func(a, b string) int { return cmp.Compare(len(b), len(a)) })
	return kept
}

func scrubRoot(text, root string) string {
	var out strings.Builder
	last := 0
	for _, at := range boundedStarts(text, root) {
		out.WriteString(text[last:at])
		last = at + len(root)
		if strings.HasPrefix(text[last:], "/") {
			last++
		} else {
			out.WriteString(".")
		}
	}
	out.WriteString(text[last:])
	return out.String()
}

func boundedStarts(text, root string) []int {
	var starts []int
	for from := 0; from < len(text); from++ {
		at := strings.Index(text[from:], root)
		if at < 0 {
			break
		}
		from += at
		if startsPath(text[:from]) && endsPath(text[from+len(root):]) {
			starts = append(starts, from)
			from += len(root) - 1
		}
	}
	return starts
}

func startsPath(before string) bool {
	last, _ := utf8.DecodeLastRuneInString(before)
	return before == "" || !pathRune(last)
}

func endsPath(after string) bool {
	first, _ := utf8.DecodeRuneInString(after)
	switch {
	case after == "" || first == '/':
		return true
	case first == '.':
		return endsPath(after[1:])
	}
	return !pathRune(first)
}

func pathRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("/.-_~+@%", r)
}
