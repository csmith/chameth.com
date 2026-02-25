package playedbgs

type Game struct {
	Name      string
	Year      int
	ImagePath string
	PlayCount int
}

type Data struct {
	Games []Game
}
