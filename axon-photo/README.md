# axon-photo

Image generation tools for LLM-powered agents.

Handles prompt merging, image storage, gallery management, and task submission for image generation pipelines.

## Install

```
go get github.com/benaskins/axon-photo@latest
```

Requires Go 1.24+.

## Usage

```go
cfg := &photo.Config{
    PromptMerger: promptMerger,
    ImageStore:   imageStore,
    GalleryStore: galleryStore,
    MessageStore: messageStore,
}

tools := map[string]tool.ToolDef{
    "take_photo":         photo.TakePhotoTool(cfg),
    "take_private_photo": photo.TakePrivatePhotoTool(cfg),
}
```

### Key types

- `Config` — wires together prompt merging, storage, and task submission
- `TakePhotoTool()`, `TakePrivatePhotoTool()` — tool constructors for LLM agents
- `PromptMerger` — merges user prompts with style/character configuration
- `ImageStore` — local image file storage with thumbnails
- `GalleryStore` — persistence interface for gallery images
- `ImageHandler()`, `GalleryListHandler()` — HTTP handlers for serving images

## License

Apache 2.0 — see [LICENSE](LICENSE).
