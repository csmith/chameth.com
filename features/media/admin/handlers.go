package admin

import (
	"bytes"
	"context"
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

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/media/admin/templates"
	"chameth.com/chameth.com/features/pages"
	"chameth.com/chameth.com/features/poems"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/features/videogames"
)

func MediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaList, err := media.GetAllMedia(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve media", http.StatusInternalServerError)
			return
		}

		items := make([]templates.MediaItem, len(mediaList))
		for i, m := range mediaList {
			items[i] = templates.MediaItem{
				ID:               m.ID,
				OriginalFilename: m.OriginalFilename,
				ParentMediaID:    m.ParentMediaID,
				Width:            m.Width,
				Height:           m.Height,
				ContentType:      m.ContentType,
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
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			http.Error(w, "No files uploaded", http.StatusBadRequest)
			return
		}

		fileGroups := groupFilesByBaseName(files)

		for baseName, group := range fileGroups {
			if err := processMediaGroup(r.Context(), baseName, group); err != nil {
				slog.Error("Failed to process media group", "baseName", baseName, "error", err)
				http.Error(w, "Failed to process uploaded files", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/media", http.StatusSeeOther)
	}
}

type fileInfo struct {
	header *multipart.FileHeader
	ext    string
}

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

func isOriginalImage(ext string) bool {
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func isImage(ext string) bool {
	return isOriginalImage(ext) || ext == ".webp" || ext == ".avif"
}

func processMediaGroup(ctx context.Context, baseName string, files []fileInfo) error {
	var originalFile *fileInfo
	for i := range files {
		if isOriginalImage(files[i].ext) {
			originalFile = &files[i]
			break
		}
	}

	var width, height *int
	if originalFile != nil {
		w, h, err := getImageDimensions(originalFile.header)
		if err != nil {
			slog.Error("Failed to get image dimensions", "filename", originalFile.header.Filename, "error", err)
		} else {
			width, height = &w, &h
		}
	}

	var parentMediaID *int
	if originalFile != nil {
		id, err := createMediaFromFile(ctx, originalFile.header, width, height, nil)
		if err != nil {
			return err
		}
		parentMediaID = &id
		slog.Info("Created original media", "id", id, "filename", originalFile.header.Filename)
	}

	for i := range files {
		if originalFile != nil && &files[i] == originalFile {
			continue
		}

		var w, h *int
		if isImage(files[i].ext) {
			w, h = width, height
		}

		id, err := createMediaFromFile(ctx, files[i].header, w, h, parentMediaID)
		if err != nil {
			return err
		}
		slog.Info("Created media variant/file", "id", id, "filename", files[i].header.Filename, "parent_id", parentMediaID)
	}

	return nil
}

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

func createMediaFromFile(ctx context.Context, fileHeader *multipart.FileHeader, width, height, parentMediaID *int) (int, error) {
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

	return media.CreateMedia(ctx, contentType, fileHeader.Filename, data, width, height, parentMediaID)
}

func EditMediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}

		m, err := media.GetMediaByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		data := templates.EditMediaData{
			ID:               m.ID,
			OriginalFilename: m.OriginalFilename,
			ParentMediaID:    m.ParentMediaID,
			Width:            m.Width,
			Height:           m.Height,
			ContentType:      m.ContentType,
		}
		if err := templates.RenderEditMedia(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func ReplaceMediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("replacement")
		if err != nil {
			http.Error(w, "No replacement file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		var width, height *int
		if isImage(strings.ToLower(filepath.Ext(header.Filename))) {
			config, _, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil {
				slog.Error("Failed to get image dimensions", "filename", header.Filename, "error", err)
			} else {
				cw, ch := config.Width, config.Height
				width, height = &cw, &ch
			}
		}

		if err := media.UpdateMedia(r.Context(), id, contentType, header.Filename, data, width, height); err != nil {
			slog.Error("Failed to update media", "id", id, "error", err)
			http.Error(w, "Failed to update media", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/media", http.StatusSeeOther)
	}
}

func ViewMediaHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}

		m, err := media.GetMediaByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", m.ContentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", m.OriginalFilename))
		w.Write(m.Data)
	}
}

func EditMediaRelationsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		entityType := r.URL.Query().Get("entity_type")
		entityIDStr := r.URL.Query().Get("entity_id")

		if entityType == "" || entityIDStr == "" {
			http.Error(w, "Missing entity_type or entity_id", http.StatusBadRequest)
			return
		}

		entityID, err := strconv.Atoi(entityIDStr)
		if err != nil {
			http.Error(w, "Invalid entity_id", http.StatusBadRequest)
			return
		}

		var entityPath string
		switch entityType {
		case "post":
			post, err := posts.GetPostByID(r.Context(), entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = post.Path
		case "poem":
			poem, err := poems.GetPoemByID(r.Context(), entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = poem.Path
		case "snippet":
			snippet, err := snippets.GetSnippetByID(r.Context(), entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = snippet.Path
		case "staticpage":
			page, err := pages.GetStaticPageByID(r.Context(), entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = page.Path
		case "film":
			film, err := films.GetFilmByID(r.Context(), entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = fmt.Sprintf("/film-%d/", film.ID)
		case "videogame":
			game, err := videogames.GetVideoGameByID(r.Context(), entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = game.Path
		default:
			http.Error(w, "Unsupported entity type", http.StatusBadRequest)
			return
		}

		mediaRelations, err := media.GetMediaRelationsForEntity(r.Context(), entityType, entityID)
		if err != nil {
			http.Error(w, "Failed to retrieve media relations", http.StatusInternalServerError)
			return
		}

		primaryMedia := make([]templates.MediaRelationItem, 0)
		variantsByParent := make(map[int][]templates.MediaRelationItem)

		for _, rel := range mediaRelations {
			caption := ""
			if rel.Caption != nil {
				caption = *rel.Caption
			}
			description := ""
			if rel.Description != nil {
				description = *rel.Description
			}
			role := ""
			if rel.Role != nil {
				role = *rel.Role
			}

			item := templates.MediaRelationItem{
				Path:        rel.Path,
				Title:       caption,
				AltText:     description,
				Width:       rel.Width,
				Height:      rel.Height,
				Role:        role,
				ContentType: rel.ContentType,
				MediaID:     rel.MediaID,
				IsVariant:   rel.ParentMediaID != nil,
			}

			if rel.ParentMediaID == nil {
				primaryMedia = append(primaryMedia, item)
			} else {
				variantsByParent[*rel.ParentMediaID] = append(variantsByParent[*rel.ParentMediaID], item)
			}
		}

		mediaItems := make([]templates.MediaRelationItem, 0, len(mediaRelations))
		for _, primary := range primaryMedia {
			mediaItems = append(mediaItems, primary)
			if variants, exists := variantsByParent[primary.MediaID]; exists {
				mediaItems = append(mediaItems, variants...)
			}
		}

		availableMedia, err := media.GetAvailableMediaForEntity(r.Context(), entityType, entityID)
		if err != nil {
			http.Error(w, "Failed to retrieve available media", http.StatusInternalServerError)
			return
		}

		availableMediaItems := make([]templates.AvailableMediaItem, 0, len(availableMedia))
		for _, m := range availableMedia {
			availableMediaItems = append(availableMediaItems, templates.AvailableMediaItem{
				MediaID:          m.ID,
				OriginalFilename: m.OriginalFilename,
				ContentType:      m.ContentType,
				Width:            m.Width,
				Height:           m.Height,
				IsVariant:        m.ParentMediaID != nil,
			})
		}

		data := templates.EditMediaRelationsData{
			EntityType:     entityType,
			EntityID:       entityID,
			EntityPath:     entityPath,
			Media:          mediaItems,
			AvailableMedia: availableMediaItems,
		}

		if err := templates.RenderEditMediaRelations(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func UpdateMediaRelationHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		entityType := r.FormValue("entity_type")
		entityIDStr := r.FormValue("entity_id")
		mediaIDStr := r.FormValue("media_id")
		path := r.FormValue("path")
		title := r.FormValue("title")
		altText := r.FormValue("alt_text")
		role := r.FormValue("role")

		entityID, err := strconv.Atoi(entityIDStr)
		if err != nil {
			http.Error(w, "Invalid entity_id", http.StatusBadRequest)
			return
		}

		mediaID, err := strconv.Atoi(mediaIDStr)
		if err != nil {
			http.Error(w, "Invalid media_id", http.StatusBadRequest)
			return
		}

		var titlePtr, altTextPtr, rolePtr *string
		if title != "" {
			titlePtr = &title
		}
		if altText != "" {
			altTextPtr = &altText
		}
		if role != "" {
			rolePtr = &role
		}

		if err := media.UpdateMediaRelation(r.Context(), entityType, entityID, path, titlePtr, altTextPtr, rolePtr); err != nil {
			http.Error(w, "Failed to update media relation", http.StatusInternalServerError)
			return
		}

		if err := media.UpdateMediaRelationVariants(r.Context(), entityType, entityID, mediaID, titlePtr, altTextPtr); err != nil {
			http.Error(w, "Failed to update variant media relations", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/media-relations/edit?entity_type=%s&entity_id=%d", entityType, entityID), http.StatusSeeOther)
	}
}

func RemoveMediaRelationHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		entityType := r.FormValue("entity_type")
		entityIDStr := r.FormValue("entity_id")
		path := r.FormValue("path")

		entityID, err := strconv.Atoi(entityIDStr)
		if err != nil {
			http.Error(w, "Invalid entity_id", http.StatusBadRequest)
			return
		}

		if err := media.DeleteMediaRelation(r.Context(), entityType, entityID, path); err != nil {
			http.Error(w, "Failed to remove media relation", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/media-relations/edit?entity_type=%s&entity_id=%d", entityType, entityID), http.StatusSeeOther)
	}
}

func AddMediaRelationsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		entityType := r.FormValue("entity_type")
		entityIDStr := r.FormValue("entity_id")
		entityPath := r.FormValue("entity_path")
		selectedMedia := r.Form["media_ids"]

		entityID, err := strconv.Atoi(entityIDStr)
		if err != nil {
			http.Error(w, "Invalid entity_id", http.StatusBadRequest)
			return
		}

		for _, mediaIDStr := range selectedMedia {
			mediaID, err := strconv.Atoi(mediaIDStr)
			if err != nil {
				continue
			}

			m, err := media.GetMediaByID(r.Context(), mediaID)
			if err != nil {
				continue
			}

			path := entityPath + m.OriginalFilename

			var rolePtr *string
			if m.ParentMediaID != nil {
				role := "alternative"
				rolePtr = &role
			}

			if err := media.CreateMediaRelation(r.Context(), entityType, entityID, mediaID, path, nil, nil, rolePtr); err != nil {
				http.Error(w, "Failed to create media relation", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/media-relations/edit?entity_type=%s&entity_id=%d", entityType, entityID), http.StatusSeeOther)
	}
}
