package domain

import "strings"

const spanMark = "\x1f"

var textEscaper = strings.NewReplacer("\\", `\\`, "|", `\|`, "*", `\*`, "_", `\_`, "`", "\\`", "[", `\[`, "]", `\]`,
	"<", `\<`, ">", `\>`, "&", `\&`, "\r\n", " ", "\n", " ", "\r", " ")

var spanEscaper = strings.NewReplacer("|", `\|`, "\r\n", " ", "\n", " ", "\r", " ")

func escapeText(text string) string {
	return strings.TrimSpace(textEscaper.Replace(text))
}

func escapeCell(cell string) string {
	parts := strings.Split(cell, spanMark)
	for i, part := range parts {
		if i%2 == 1 {
			parts[i] = codeSpan(spanEscaper.Replace(part))
		} else {
			parts[i] = textEscaper.Replace(part)
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}

func codeSpan(content string) string {
	fence := strings.Repeat("`", longestBacktickRun(content)+1)
	if strings.HasPrefix(content, "`") || strings.HasSuffix(content, "`") {
		content = " " + content + " "
	}
	return fence + content + fence
}

func longestBacktickRun(text string) int {
	longest, run := 0, 0
	for _, r := range text {
		if r == '`' {
			run++
		} else {
			run = 0
		}
		longest = max(longest, run)
	}
	return longest
}
