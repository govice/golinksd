package main

//CardOption is used in templating to display options in a ConsoleCard
type CardOption struct {
	Label string `json:"label"`
	URL   string `json:"URL"`
}

type CardButton struct {
	Label string
	URL   string
	Class string
}

//ConsoleCard is in templating to display a card on the console
type ConsoleCard struct {
	Title   string        `json:"title"`
	Options []*CardOption `json:"options"`
	Buttons []*CardButton `json:"buttons"`
}
