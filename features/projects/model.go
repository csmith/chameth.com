package projects

type ProjectSection struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Sort        int    `db:"sort"`
	Description string `db:"description"`
}

type Project struct {
	ID          int    `db:"id"`
	Section     int    `db:"section"`
	Name        string `db:"name"`
	Icon        string `db:"icon"`
	Pinned      bool   `db:"pinned"`
	Description string `db:"description"`
	Published   bool   `db:"published"`
}
