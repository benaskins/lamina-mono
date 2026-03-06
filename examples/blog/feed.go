package main

import (
	"log/slog"
	"net/http"

	"github.com/benaskins/axon"
)

// FeedHandler renders an Atom feed of recent posts.
func (b *Blog) FeedHandler(w http.ResponseWriter, r *http.Request) {
	posts := b.posts
	if len(posts) > 20 {
		posts = posts[:20]
	}

	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	err := b.templates.ExecuteTemplate(w, "feed.xml", map[string]any{
		"Site":  b.config,
		"Posts": posts,
	})
	if err != nil {
		slog.Error("render feed", "error", err)
		axon.WriteError(w, http.StatusInternalServerError, "render failed")
	}
}
