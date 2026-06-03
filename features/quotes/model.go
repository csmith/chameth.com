package quotes

type Quote struct {
	ID     int    `db:"id"`
	Text   string `db:"text"`
	Author string `db:"author"`
}
