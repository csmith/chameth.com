package feeds

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/posts"
	parenttemplates "chameth.com/chameth.com/templates"
)

//go:embed builder.html.gotpl
var builderTemplates embed.FS

var builderTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))

	template.Must(t.ParseFS(builderTemplates, "builder.html.gotpl"))
	return t
}()

const builderFeedPrefix = "/feeds/posts/build/"

// builderPostLimit caps how many posts the similarity query returns when
// classifying posts for the builder page. It is deliberately far higher than
// the feed's own limit: the page classifies every post against the
// like/unlike criteria rather than showing just the most recent matches.
const builderPostLimit = 500

type BuilderData struct {
	parenttemplates.PageData
	Sections []builderSection
	FeedUrl  string
}

type builderSection struct {
	Title string
	Posts []builderPost
}

type builderPost struct {
	Title   string
	Url     string
	Date    string
	Pinned  bool
	Actions []builderAction
}

type builderAction struct {
	Label string
	Url   string
}

func handleRelatedPostsBuilder(w http.ResponseWriter, r *http.Request) {
	likes, unlikes, ok := parseRelatedFeedParams(strings.TrimPrefix(r.URL.Path, builderFeedPrefix))
	if !ok {
		http.NotFound(w, r)
		return
	}

	canonicalPath := builderFeedPrefix + relatedFeedParams(likes, unlikes)
	if r.URL.Path != canonicalPath {
		w.Header().Set("Location", canonicalPath)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	allPosts, err := posts.GetAllPosts(r.Context())
	if err != nil {
		slog.Error("Failed to get posts for feed builder", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var included []posts.Post
	if len(likes) > 0 || len(unlikes) > 0 {
		included, err = posts.GetRecentPostsBySimilarity(r.Context(), likes, unlikes, builderPostLimit)
		if err != nil {
			slog.Error("Failed to score posts for feed builder", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data := BuilderData{
		PageData: content.CreatePageData(r.Context(), "Build a post feed", builderFeedPrefix, parenttemplates.OpenGraphHeaders{}),
		Sections: builderSections(allPosts, included, likes, unlikes),
	}
	if len(likes) > 0 || len(unlikes) > 0 {
		data.FeedUrl = relatedFeedPath(likes, unlikes)
	}
	data.Robots = "noindex, nofollow"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = builderTemplate.Execute(w, data)
	if err != nil {
		slog.Error("Failed to render feed builder template", "error", err)
	}
}

// builderSections splits every published post into the builder's two
// sections. The explicitly chosen likes and unlikes ("seeds") are listed at
// the top of their section and flagged Pinned so the template can annotate
// them. Unliked posts never match the feed criteria (their own embedding
// zeroes their score) so they always fall under "Posts not included". Empty
// sections are omitted.
func builderSections(allPosts []posts.PostMetadata, included []posts.Post, likes, unlikes []string) []builderSection {
	likeSet := make(map[string]bool, len(likes))
	for _, slug := range likes {
		likeSet[slug] = true
	}

	unlikeSet := make(map[string]bool, len(unlikes))
	for _, slug := range unlikes {
		unlikeSet[slug] = true
	}

	includedIDs := make(map[int]bool, len(included))
	for _, post := range included {
		includedIDs[post.ID] = true
	}

	var contentsSeeds, contentsPosts, notIncludedSeeds, notIncludedPosts []builderPost
	for _, post := range allPosts {
		slug := strings.Trim(post.Path, "/")
		entry := builderPost{
			Title:   post.Title,
			Url:     post.Path,
			Date:    post.Date.Format("Jan 2, 2006"),
			Pinned:  likeSet[slug] || unlikeSet[slug],
			Actions: builderActions(post.Path, likes, unlikes),
		}

		switch {
		case likeSet[slug]:
			contentsSeeds = append(contentsSeeds, entry)
		case unlikeSet[slug]:
			notIncludedSeeds = append(notIncludedSeeds, entry)
		case includedIDs[post.ID]:
			contentsPosts = append(contentsPosts, entry)
		default:
			notIncludedPosts = append(notIncludedPosts, entry)
		}
	}

	sections := []builderSection{
		{Title: "Feed contents", Posts: append(contentsSeeds, contentsPosts...)},
		{Title: "Posts not included", Posts: append(notIncludedSeeds, notIncludedPosts...)},
	}

	var result []builderSection
	for _, section := range sections {
		if len(section.Posts) > 0 {
			result = append(result, section)
		}
	}
	return result
}

// builderActions returns the links that add/remove a post's slug to/from the
// like and unlike lists. Seeded posts get only the link that removes them
// from the list they are in; unseeded posts get links to add them to either
// list. Posts whose path cannot be expressed as a feed slug get no actions.
func builderActions(path string, likes, unlikes []string) []builderAction {
	slug := strings.Trim(path, "/")
	if !relatedFeedSlugRegex.MatchString(slug) {
		return nil
	}

	// Normalisation guarantees a slug is never in both lists.
	if slices.Contains(likes, slug) {
		return []builderAction{{
			Label: "Don't include posts like this",
			Url:   builderPath(withoutSlug(likes, slug), unlikes),
		}}
	}
	if slices.Contains(unlikes, slug) {
		return []builderAction{{
			Label: "Don't exclude posts like this",
			Url:   builderPath(likes, withoutSlug(unlikes, slug)),
		}}
	}

	var actions []builderAction
	if len(likes) < maxFeedSlugs {
		actions = append(actions, builderAction{
			Label: "Include posts like this",
			Url:   builderPath(withSlug(likes, slug), unlikes),
		})
	}
	if len(unlikes) < maxFeedSlugs {
		actions = append(actions, builderAction{
			Label: "Exclude posts like this",
			Url:   builderPath(likes, withSlug(unlikes, slug)),
		})
	}
	return actions
}

func builderPath(likes, unlikes []string) string {
	return builderFeedPrefix + relatedFeedParams(likes, unlikes)
}

func withSlug(slugs []string, slug string) []string {
	if slices.Contains(slugs, slug) {
		return slugs
	}
	return sortAndDedupe(append(slices.Clone(slugs), slug))
}

func withoutSlug(slugs []string, slug string) []string {
	return sortAndDedupe(slices.DeleteFunc(slices.Clone(slugs), func(s string) bool { return s == slug }))
}
