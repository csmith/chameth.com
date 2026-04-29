package goimports

type GoImport struct {
	ID        int    `db:"id"`
	Path      string `db:"path"`
	VCS       string `db:"vcs"`
	RepoURL   string `db:"repo_url"`
	Published bool   `db:"published"`
}
