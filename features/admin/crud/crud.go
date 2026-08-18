// Package crud provides handlers for the standard admin CRUD flow: a list
// page (optionally split into drafts and published items), an edit page, a
// create action, and a form-driven update action.
package crud

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
	"github.com/csmith/aca"
)

// Routes holds the handlers for the standard CRUD route set. See Register.
type Routes struct {
	List   http.HandlerFunc
	Create http.HandlerFunc
	Edit   http.HandlerFunc
	Update http.HandlerFunc
}

// Register registers the standard CRUD routes under basePath:
// GET/POST basePath, and GET/POST basePath/edit/{id}.
func Register(mux *http.ServeMux, basePath string, routes Routes) {
	mux.HandleFunc("GET "+basePath, routes.List)
	mux.HandleFunc("POST "+basePath, routes.Create)
	mux.HandleFunc("GET "+basePath+"/edit/{id}", routes.Edit)
	mux.HandleFunc("POST "+basePath+"/edit/{id}", routes.Update)
}

// List returns a handler that fetches draft and published items, maps them
// to summaries, and renders the list page.
func List[M, S any](
	entity string,
	fetch func(context.Context) ([]M, []M, error),
	toSummary func(M) S,
	render func(http.ResponseWriter, admintemplates.ListData[S]) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		draftItems, allItems, err := fetch(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve "+entity+"s", http.StatusInternalServerError)
			return
		}

		data := admintemplates.ListData[S]{
			Drafts: mapAll(draftItems, toSummary),
			Items:  mapAll(allItems, toSummary),
		}
		if err := render(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

// DraftsAndAll combines the typical separate draft and published item
// queries into the single fetch function List expects.
func DraftsAndAll[M any](drafts, all func(context.Context) ([]M, error)) func(context.Context) ([]M, []M, error) {
	return func(ctx context.Context) ([]M, []M, error) {
		draftItems, err := drafts(ctx)
		if err != nil {
			return nil, nil, err
		}
		allItems, err := all(ctx)
		if err != nil {
			return nil, nil, err
		}
		return draftItems, allItems, nil
	}
}

// AllItems adapts a plain fetch of all items to the fetch function List
// expects, with no draft items.
func AllItems[M any](all func(context.Context) ([]M, error)) func(context.Context) ([]M, []M, error) {
	return func(ctx context.Context) ([]M, []M, error) {
		items, err := all(ctx)
		return nil, items, err
	}
}

// Edit returns a handler that fetches a single item by its path ID and
// renders the edit page for it.
func Edit[T, E any](
	entity string,
	fetchOne func(context.Context, int) (T, error),
	toEditData func(context.Context, T) (E, error),
	render func(http.ResponseWriter, E) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			http.Error(w, "Invalid "+entity+" ID", http.StatusBadRequest)
			return
		}

		item, err := fetchOne(r.Context(), id)
		if err != nil {
			http.Error(w, capitalise(entity)+" not found", http.StatusNotFound)
			return
		}

		data, err := toEditData(r.Context(), item)
		if err != nil {
			http.Error(w, "Failed to retrieve "+entity, http.StatusInternalServerError)
			return
		}

		if err := render(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

// Create returns a handler that creates a new item and redirects to its
// edit page.
func Create(entity, basePath string, create func(*http.Request) (int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := create(r)
		if err != nil {
			http.Error(w, "Failed to create "+entity, http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, editPath(basePath, id), http.StatusSeeOther)
	}
}

// GeneratePath returns a create function that picks a new aca-generated name
// and creates an item at the path formed from that name, e.g. "/paste/<name>/".
func GeneratePath(pathFormat string, create func(context.Context, string, string) (int, error)) func(*http.Request) (int, error) {
	return func(r *http.Request) (int, error) {
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			return 0, err
		}
		name := gen.Generate()
		return create(r.Context(), fmt.Sprintf(pathFormat, name), name)
	}
}

// GenerateName returns a create function that picks a new aca-generated name
// for items that have no path of their own.
func GenerateName(create func(context.Context, string) (int, error)) func(*http.Request) (int, error) {
	return func(r *http.Request) (int, error) {
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			return 0, err
		}
		return create(r.Context(), gen.Generate())
	}
}

// Update returns a handler that applies a parsed form submission to an
// existing item and redirects back to its edit page.
func Update(entity, basePath string, apply func(context.Context, int, url.Values) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			http.Error(w, "Invalid "+entity+" ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		if err := apply(r.Context(), id, r.PostForm); err != nil {
			http.Error(w, "Failed to update "+entity, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, editPath(basePath, id), http.StatusSeeOther)
	}
}

func mapAll[M, S any](items []M, fn func(M) S) []S {
	summaries := make([]S, len(items))
	for i, item := range items {
		summaries[i] = fn(item)
	}
	return summaries
}

func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func editPath(basePath string, id int) string {
	return fmt.Sprintf("%s/edit/%d", basePath, id)
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
