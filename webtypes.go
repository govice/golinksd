package main

//CardOption is used in templating to display options in a ConsoleCard
type CardOption struct {
	Label string `json:"label"`
	URL   string `json:"URL"`
}

//ConsoleCard is in templating to display a card on the console
type ConsoleCard struct {
	Title   string       `json:"title"`
	Options []CardOption `json:"options"`
}
