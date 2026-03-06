package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/benaskins/axon"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"gopkg.in/yaml.v3"
)

// SiteConfig holds top-level blog configuration.
type SiteConfig struct {
	Title   string
	BaseURL string
}

// Post represents a single blog post loaded from a markdown file.
type Post struct {
	Slug    string
	Title   string
	Date    time.Time
	Tags    []string
	Summary string
	Content template.HTML // rendered markdown
}

// frontMatter is the YAML block at the top of each markdown file.
type frontMatter struct {
	Title   string   `yaml:"title"`
	Date    string   `yaml:"date"`
	Tags    []string `yaml:"tags"`
	Summary string   `yaml:"summary"`
}

// Blog loads posts from disk and renders them via templates.
type Blog struct {
	posts     []Post
	postIndex map[string]*Post
	templates *template.Template
	config    SiteConfig
	md        goldmark.Markdown
}

// NewBlog loads all posts from contentDir/posts/ and parses templates
// from the embedded filesystem.
func NewBlog(contentDir string, templateFS embed.FS, config SiteConfig) (*Blog, error) {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2 January 2006")
		},
		"isoDate": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html", "templates/*.xml")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
		),
	)

	b := &Blog{
		postIndex: make(map[string]*Post),
		templates: tmpl,
		config:    config,
		md:        md,
	}

	if err := b.loadPosts(filepath.Join(contentDir, "posts")); err != nil {
		return nil, err
	}

	return b, nil
}

// loadPosts reads all .md files from dir, parses front matter and
// renders markdown to HTML.
func (b *Blog) loadPosts(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read posts dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		post, err := b.loadPost(filepath.Join(dir, entry.Name()))
		if err != nil {
			slog.Warn("skipping post", "file", entry.Name(), "error", err)
			continue
		}

		b.posts = append(b.posts, post)
	}

	// newest first
	sort.Slice(b.posts, func(i, j int) bool {
		return b.posts[i].Date.After(b.posts[j].Date)
	})

	for i := range b.posts {
		b.postIndex[b.posts[i].Slug] = &b.posts[i]
	}

	return nil
}

// loadPost reads a single markdown file, splits front matter from body,
// and renders the body to HTML.
func (b *Blog) loadPost(path string) (Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Post{}, err
	}

	fm, body, err := parseFrontMatter(data)
	if err != nil {
		return Post{}, fmt.Errorf("front matter: %w", err)
	}

	date, err := time.Parse("2006-01-02", fm.Date)
	if err != nil {
		return Post{}, fmt.Errorf("parse date %q: %w", fm.Date, err)
	}

	var buf bytes.Buffer
	if err := b.md.Convert(body, &buf); err != nil {
		return Post{}, fmt.Errorf("render markdown: %w", err)
	}

	// derive slug from filename: 2024-01-15-my-post.md → my-post
	slug := strings.TrimSuffix(filepath.Base(path), ".md")
	if parts := strings.SplitN(slug, "-", 4); len(parts) == 4 {
		slug = parts[3]
	}

	return Post{
		Slug:    slug,
		Title:   fm.Title,
		Date:    date,
		Tags:    fm.Tags,
		Summary: fm.Summary,
		Content: template.HTML(buf.String()),
	}, nil
}

// parseFrontMatter splits a document on --- delimiters and unmarshals
// the YAML header.
func parseFrontMatter(data []byte) (frontMatter, []byte, error) {
	const delimiter = "---"

	s := string(data)
	s = strings.TrimLeft(s, "\n\r")

	if !strings.HasPrefix(s, delimiter) {
		return frontMatter{}, data, fmt.Errorf("missing front matter")
	}

	s = s[len(delimiter):]
	end := strings.Index(s, "\n"+delimiter)
	if end < 0 {
		return frontMatter{}, data, fmt.Errorf("unclosed front matter")
	}

	var fm frontMatter
	if err := yaml.Unmarshal([]byte(s[:end]), &fm); err != nil {
		return frontMatter{}, nil, err
	}

	body := []byte(s[end+len("\n"+delimiter):])
	body = bytes.TrimLeft(body, "\n\r")

	return fm, body, nil
}

// IndexHandler renders the post listing page.
func (b *Blog) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := b.templates.ExecuteTemplate(w, "index.html", map[string]any{
		"Title": b.config.Title,
		"Posts": b.posts,
	})
	if err != nil {
		slog.Error("render index", "error", err)
		axon.WriteError(w, http.StatusInternalServerError, "render failed")
	}
}

// PostHandler renders a single post by slug.
func (b *Blog) PostHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, ok := b.postIndex[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := b.templates.ExecuteTemplate(w, "post.html", map[string]any{
		"Title": post.Title + " — " + b.config.Title,
		"Post":  post,
		"Site":  b.config,
	})
	if err != nil {
		slog.Error("render post", "error", err, "slug", slug)
		axon.WriteError(w, http.StatusInternalServerError, "render failed")
	}
}
