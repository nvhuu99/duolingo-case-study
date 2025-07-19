package models

import (
	"time"
)

const USER_ID_KEY = "user_id"

type User struct {
	Id              string         `json:"user_id" bson:"user_id"`
	Lastname        string         `json:"lastname" bson:"lastname"`
	Firstname       string         `json:"firstname" bson:"firstname"`
	Username        string         `json:"username" bson:"username"`
	Email           string         `json:"email" bson:"email"`
	Campaigns       []string       `json:"campaigns" bson:"campaigns"`
	Devices         []*UserDevice  `json:"user_devices" bson:"user_devices"`
	NativeLanguage  NativeLanguage `json:"native_lan_enum" bson:"native_lan_enum"`
	Membership      Membership     `json:"membership_enum" bson:"membership_enum"`
	EmailVerifiedAt time.Time      `json:"email_verified_at" bson:"email_verified_at,omitempty"`
}

func (u *User) Equal(target *User) bool {
	if target == nil {
		return false
	}

	t1 := u.EmailVerifiedAt.UTC().Format("2006-01-02 15:04:05")
	t2 := target.EmailVerifiedAt.UTC().Format("2006-01-02 15:04:05")
	if u.Id != target.Id ||
		u.Lastname != target.Lastname ||
		u.Firstname != target.Firstname ||
		u.Username != target.Username ||
		u.Email != target.Email ||
		u.NativeLanguage != target.NativeLanguage ||
		u.Membership != target.Membership ||
		t1 != t2 {
		return false
	}

	if len(u.Campaigns) != len(target.Campaigns) {
		return false
	}
	for i := range u.Campaigns {
		if u.Campaigns[i] != target.Campaigns[i] {
			return false
		}
	}

	if len(u.Devices) != len(target.Devices) {
		return false
	}
	for i := range u.Devices {
		if *u.Devices[i] != *target.Devices[i] {
			return false
		}
	}

	return true
}
