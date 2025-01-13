package model

type InputMessage struct {
	Id        string	`json:"id"`
	Content   string	`json:"content"`
	IsRelayed bool		`json:"isRelayed"`	// Is this a relayed one or the original message
	Campaign  string	`json:"campaign"`
}
