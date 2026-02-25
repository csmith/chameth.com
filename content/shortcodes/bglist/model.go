package bglist

type Game struct {
	Position   int
	Name       string
	Year       int
	ImagePath  string
	PlayCount  int
	LastPlayed string
}

type Data struct {
	Games []Game
}
