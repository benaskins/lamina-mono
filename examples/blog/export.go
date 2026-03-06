package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Export renders the entire site to static files under outDir.
func (b *Blog) Export(outDir string, staticFS fs.FS) error {
	// index
	var buf bytes.Buffer
	if err := b.templates.ExecuteTemplate(&buf, "index.html", map[string]any{
		"Title": b.config.Title,
		"Posts": b.posts,
	}); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	if err := writeFile(filepath.Join(outDir, "index.html"), buf.Bytes()); err != nil {
		return err
	}

	// individual posts
	for i := range b.posts {
		post := &b.posts[i]
		buf.Reset()
		if err := b.templates.ExecuteTemplate(&buf, "post.html", map[string]any{
			"Title": post.Title + " — " + b.config.Title,
			"Post":  post,
			"Site":  b.config,
		}); err != nil {
			return fmt.Errorf("render post %s: %w", post.Slug, err)
		}
		if err := writeFile(filepath.Join(outDir, "posts", post.Slug, "index.html"), buf.Bytes()); err != nil {
			return err
		}
	}

	// feed
	posts := b.posts
	if len(posts) > 20 {
		posts = posts[:20]
	}
	buf.Reset()
	if err := b.templates.ExecuteTemplate(&buf, "feed.xml", map[string]any{
		"Site":  b.config,
		"Posts": posts,
	}); err != nil {
		return fmt.Errorf("render feed: %w", err)
	}
	if err := writeFile(filepath.Join(outDir, "feed.xml"), buf.Bytes()); err != nil {
		return err
	}

	// static assets
	return fs.WalkDir(staticFS, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(staticFS, path)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(outDir, path), data)
	})
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
