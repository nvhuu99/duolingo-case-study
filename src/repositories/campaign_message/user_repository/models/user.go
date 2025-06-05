package models

import "time"

type User struct {
	Id              string
	Lastname        string
	Firstname       string
	Username        string
	Email           string
	Campaigns       []string
	DeviceTokens    []string
	NativeLanguage  NativeLanguage
	Membership      Membership
	EmailVerifiedAt time.Time
}
