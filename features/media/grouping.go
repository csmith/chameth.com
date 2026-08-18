package media

// GroupedMedia is a primary media item with its variants attached.
type GroupedMedia struct {
	Path        string
	Title       string
	AltText     string
	Width       *int
	Height      *int
	Role        string
	ContentType string
	MediaID     int
	Variants    []GroupedMediaVariant
}

// GroupedMediaVariant is a media item that is a variant of a parent.
type GroupedMediaVariant struct {
	MediaID     int
	ContentType string
	Width       *int
	Height      *int
}

// GroupByPrimary flattens media relations into primary items, attaching
// relations that have a parent media item to that parent's Variants.
func GroupByPrimary(relations []MediaRelationWithDetails) []GroupedMedia {
	primaries := make(map[int]*GroupedMedia)
	var primaryIDs []int

	for _, rel := range relations {
		if rel.ParentMediaID != nil {
			continue
		}
		if _, exists := primaries[rel.MediaID]; exists {
			continue
		}

		primaryIDs = append(primaryIDs, rel.MediaID)
		primaries[rel.MediaID] = &GroupedMedia{
			Path:        rel.Path,
			Title:       valueOrEmpty(rel.Caption),
			AltText:     valueOrEmpty(rel.Description),
			Width:       rel.Width,
			Height:      rel.Height,
			Role:        valueOrEmpty(rel.Role),
			ContentType: rel.ContentType,
			MediaID:     rel.MediaID,
			Variants:    []GroupedMediaVariant{},
		}
	}

	for _, rel := range relations {
		if rel.ParentMediaID == nil {
			continue
		}
		if parent, exists := primaries[*rel.ParentMediaID]; exists {
			parent.Variants = append(parent.Variants, GroupedMediaVariant{
				MediaID:     rel.MediaID,
				ContentType: rel.ContentType,
				Width:       rel.Width,
				Height:      rel.Height,
			})
		}
	}

	grouped := make([]GroupedMedia, 0, len(primaryIDs))
	for _, id := range primaryIDs {
		grouped = append(grouped, *primaries[id])
	}
	return grouped
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
