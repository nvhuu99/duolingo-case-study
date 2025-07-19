package models

type UserDevice struct {
	Platform string `json:"platform" bson:"platform"`
	Token    string `json:"token" bson:"token"`
}
