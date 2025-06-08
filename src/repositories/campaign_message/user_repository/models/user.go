package models

import "time"

const USER_ID_KEY = "user_id"

type User struct {
	Id              string         `json:"user_id" bson:"user_id"`
	Lastname        string         `json:"lastname" bson:"lastname"`
	Firstname       string         `json:"firstname" bson:"firstname"`
	Username        string         `json:"username" bson:"username"`
	Email           string         `json:"email" bson:"email"`
	Campaigns       []string       `json:"campaigns" bson:"campaigns"`
	DeviceTokens    []string       `json:"device_tokens" bson:"device_tokens"`
	NativeLanguage  NativeLanguage `json:"native_lan_enum" bson:"native_lan_enum"`
	Membership      Membership     `json:"membership_enum" bson:"membership_enum"`
	EmailVerifiedAt time.Time      `json:"email_verified_at" bson:"email_verified_at"`
}
