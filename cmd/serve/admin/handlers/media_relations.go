package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

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

		// Get entity path based on type
		var entityPath string
		switch entityType {
		case "post":
			post, err := db.GetPostByID(entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = post.Path
		case "poem":
			poem, err := db.GetPoemByID(entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = poem.Path
		case "snippet":
			snippet, err := db.GetSnippetByID(entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = snippet.Path
		case "staticpage":
			page, err := db.GetStaticPageByID(entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = page.Path
		case "film":
			film, err := db.GetFilmByID(entityID)
			if err != nil {
				http.Error(w, "Entity not found", http.StatusNotFound)
				return
			}
			entityPath = fmt.Sprintf("/film-%d/", film.ID)
		default:
			http.Error(w, "Unsupported entity type", http.StatusBadRequest)
			return
		}

		// Fetch media relations for this entity
		mediaRelations, err := db.GetMediaRelationsForEntity(entityType, entityID)
		if err != nil {
			http.Error(w, "Failed to retrieve media relations", http.StatusInternalServerError)
			return
		}

		// Separate primary media and variants
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
				// Primary media
				primaryMedia = append(primaryMedia, item)
			} else {
				// Variant - group by parent
				variantsByParent[*rel.ParentMediaID] = append(variantsByParent[*rel.ParentMediaID], item)
			}
		}

		// Build final list with variants immediately after their parent
		mediaItems := make([]templates.MediaRelationItem, 0, len(mediaRelations))
		for _, primary := range primaryMedia {
			mediaItems = append(mediaItems, primary)
			// Add all variants for this parent
			if variants, exists := variantsByParent[primary.MediaID]; exists {
				mediaItems = append(mediaItems, variants...)
			}
		}

		// Fetch available media (not yet attached to this entity)
		availableMedia, err := db.GetAvailableMediaForEntity(entityType, entityID)
		if err != nil {
			http.Error(w, "Failed to retrieve available media", http.StatusInternalServerError)
			return
		}

		// Convert to template format
		availableMediaItems := make([]templates.AvailableMediaItem, 0, len(availableMedia))
		for _, media := range availableMedia {
			availableMediaItems = append(availableMediaItems, templates.AvailableMediaItem{
				MediaID:          media.ID,
				OriginalFilename: media.OriginalFilename,
				ContentType:      media.ContentType,
				Width:            media.Width,
				Height:           media.Height,
				IsVariant:        media.ParentMediaID != nil,
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

		// Convert empty strings to nil
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

		// Update the primary media relation
		if err := db.UpdateMediaRelation(entityType, entityID, path, titlePtr, altTextPtr, rolePtr); err != nil {
			http.Error(w, "Failed to update media relation", http.StatusInternalServerError)
			return
		}

		// Also update all variants with the same title and alt text
		// (This will do nothing if the media has no variants)
		if err := db.UpdateMediaRelationVariants(entityType, entityID, mediaID, titlePtr, altTextPtr); err != nil {
			http.Error(w, "Failed to update variant media relations", http.StatusInternalServerError)
			return
		}

		// Redirect back to the edit page
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

		if err := db.DeleteMediaRelation(entityType, entityID, path); err != nil {
			http.Error(w, "Failed to remove media relation", http.StatusInternalServerError)
			return
		}

		// Redirect back to the edit page
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
		selectedMedia := r.Form["media_ids"] // Array of media IDs

		entityID, err := strconv.Atoi(entityIDStr)
		if err != nil {
			http.Error(w, "Invalid entity_id", http.StatusBadRequest)
			return
		}

		// Create relations for each selected media
		for _, mediaIDStr := range selectedMedia {
			mediaID, err := strconv.Atoi(mediaIDStr)
			if err != nil {
				continue // Skip invalid IDs
			}

			// Get media metadata to get filename and check if it's a variant
			media, err := db.GetMediaByID(mediaID)
			if err != nil {
				continue // Skip if media not found
			}

			// Generate path: entity_path + filename
			path := entityPath + media.OriginalFilename

			// Determine role: variants get "alternative", others get empty
			var rolePtr *string
			if media.ParentMediaID != nil {
				role := "alternative"
				rolePtr = &role
			}

			// Create the media relation
			if err := db.CreateMediaRelation(entityType, entityID, mediaID, path, nil, nil, rolePtr); err != nil {
				http.Error(w, "Failed to create media relation", http.StatusInternalServerError)
				return
			}
		}

		// Redirect back to the edit page
		http.Redirect(w, r, fmt.Sprintf("/media-relations/edit?entity_type=%s&entity_id=%d", entityType, entityID), http.StatusSeeOther)
	}
}
