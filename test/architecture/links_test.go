package architecture

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var markdownLink = regexp.MustCompile(`\]\(([^)#\s]+)(?:#[^)]*)?\)`)

// TestDocumentationLinksResolve keeps the public documentation, the root
// pages and the agent context pointing at files that exist: a moved test or
// module must not leave a dead evidence link behind.
func TestDocumentationLinksResolve(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	pages := []string{"README.md", "AGENTS.md", "CONTRIBUTING.md"}
	for _, directory := range []string{"documentation", ".agents"} {
		err := fs.WalkDir(root.FS(), directory, func(path string, entry fs.DirEntry, err error) error {
			if err == nil && !entry.IsDir() && strings.HasSuffix(path, ".md") {
				pages = append(pages, path)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, page := range pages {
		raw, err := root.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range markdownLink.FindAllStringSubmatch(string(raw), -1) {
			target := match[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
				continue
			}
			if _, err := root.Stat(filepath.Join(filepath.Dir(page), target)); err != nil {
				t.Errorf("%s links to %s, which does not resolve", page, target)
			}
		}
	}
}
