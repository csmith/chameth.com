package handlers

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/csmith/chameth.com/cmd/serve/admin/templates"
	"github.com/csmith/chameth.com/cmd/serve/db"
)

func MediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaList, err := db.GetAllMedia()
		if err != nil {
			http.Error(w, "Failed to retrieve media", http.StatusInternalServerError)
			return
		}

		// Convert db.Media to templates.MediaItem
		items := make([]templates.MediaItem, len(mediaList))
		for i, media := range mediaList {
			items[i] = templates.MediaItem{
				ID:               media.ID,
				OriginalFilename: media.OriginalFilename,
				ParentMediaID:    media.ParentMediaID,
				Width:            media.Width,
				Height:           media.Height,
				ContentType:      media.ContentType,
			}
		}

		data := templates.MediaData{
			MediaItems: items,
		}
		if err := templates.RenderMedia(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func UploadMediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form (32MB max)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			http.Error(w, "No files uploaded", http.StatusBadRequest)
			return
		}

		// Group files by base name
		fileGroups := groupFilesByBaseName(files)

		// Process each group
		for baseName, group := range fileGroups {
			if err := processMediaGroup(baseName, group); err != nil {
				slog.Error("Failed to process media group", "baseName", baseName, "error", err)
				http.Error(w, "Failed to process uploaded files", http.StatusInternalServerError)
				return
			}
		}

		// Redirect back to media page
		http.Redirect(w, r, "/media", http.StatusSeeOther)
	}
}

type fileInfo struct {
	header *multipart.FileHeader
	ext    string
}

// groupFilesByBaseName groups files by their base name (without extension)
func groupFilesByBaseName(files []*multipart.FileHeader) map[string][]fileInfo {
	groups := make(map[string][]fileInfo)
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		baseName := strings.TrimSuffix(file.Filename, ext)
		groups[baseName] = append(groups[baseName], fileInfo{
			header: file,
			ext:    ext,
		})
	}
	return groups
}

// isOriginalImage returns true if the extension indicates an original image format
func isOriginalImage(ext string) bool {
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

// isImage returns true if the extension indicates an image format
func isImage(ext string) bool {
	return isOriginalImage(ext) || ext == ".webp" || ext == ".avif"
}

// processMediaGroup processes a group of related files (same base name)
func processMediaGroup(baseName string, files []fileInfo) error {
	// Find the original image (png/jpg/jpeg) if present
	var originalFile *fileInfo
	for i := range files {
		if isOriginalImage(files[i].ext) {
			originalFile = &files[i]
			break
		}
	}

	// If we have an original image, extract its dimensions
	var width, height *int
	if originalFile != nil {
		w, h, err := getImageDimensions(originalFile.header)
		if err != nil {
			slog.Error("Failed to get image dimensions", "filename", originalFile.header.Filename, "error", err)
		} else {
			width, height = &w, &h
		}
	}

	// First pass: create the original image (if present)
	var parentMediaID *int
	if originalFile != nil {
		id, err := createMediaFromFile(originalFile.header, width, height, nil)
		if err != nil {
			return err
		}
		parentMediaID = &id
		slog.Info("Created original media", "id", id, "filename", originalFile.header.Filename)
	}

	// Second pass: create all other files (variants and non-images)
	for i := range files {
		if originalFile != nil && &files[i] == originalFile {
			continue // Skip the original, already processed
		}

		// For image variants, use the parent's dimensions; for non-images, use nil
		var w, h *int
		if isImage(files[i].ext) {
			w, h = width, height
		}

		id, err := createMediaFromFile(files[i].header, w, h, parentMediaID)
		if err != nil {
			return err
		}
		slog.Info("Created media variant/file", "id", id, "filename", files[i].header.Filename, "parent_id", parentMediaID)
	}

	return nil
}

// getImageDimensions reads an image file and returns its width and height
func getImageDimensions(fileHeader *multipart.FileHeader) (int, int, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return 0, 0, err
	}

	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

// createMediaFromFile creates a media entry from a multipart file
func createMediaFromFile(fileHeader *multipart.FileHeader, width, height, parentMediaID *int) (int, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return 0, err
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return db.CreateMedia(contentType, fileHeader.Filename, data, width, height, parentMediaID)
}

func ViewMediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}

		media, err := db.GetMediaByID(id)
		if err != nil {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", media.ContentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", media.OriginalFilename))
		w.Write(media.Data)
	}
}
