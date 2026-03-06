// Example: a custom blog built on axon.
//
// Markdown posts live in content/posts/ with YAML front matter.
// Templates are embedded at compile time; content is loaded from disk
// at startup so you can add posts without recompiling.
//
// Run as a server:
//
//	go run . -content ./content
//
// Export to static files for Cloudflare Pages:
//
//	go run . -content ./content -export ./dist
//	npx wrangler pages deploy ./dist --project-name=my-blog
//
// Then open http://localhost:8080
package main

import (
	"embed"
	"flag"
	"log/slog"
	"net/http"

	"github.com/benaskins/axon"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templateFS embed.FS

func main() {
	contentDir := flag.String("content", "./content", "path to content directory")
	title := flag.String("title", "blog", "site title")
	baseURL := flag.String("base-url", "http://localhost:8080", "base URL for feeds")
	port := flag.String("port", "8080", "port to listen on")
	exportDir := flag.String("export", "", "export static site to directory and exit")
	flag.Parse()

	blog, err := NewBlog(*contentDir, templateFS, SiteConfig{
		Title:   *title,
		BaseURL: *baseURL,
	})
	if err != nil {
		slog.Error("failed to load blog", "error", err)
		return
	}

	if *exportDir != "" {
		if err := blog.Export(*exportDir, staticFS); err != nil {
			slog.Error("export failed", "error", err)
			return
		}
		slog.Info("exported site", "dir", *exportDir, "posts", len(blog.posts))
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /posts/{slug}", blog.PostHandler)
	mux.HandleFunc("GET /", blog.IndexHandler)
	mux.HandleFunc("GET /feed.xml", blog.FeedHandler)

	mux.Handle("GET /static/", http.FileServerFS(staticFS))

	slog.Info("starting blog", "port", *port, "posts", len(blog.posts))
	axon.ListenAndServe(*port, axon.StandardMiddleware(mux))
}
