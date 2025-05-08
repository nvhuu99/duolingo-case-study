package campaign_user

import "time"

type Membership string

const (
	Premium    Membership = "premium"
	Subscriber Membership = "subscriber"
	FreeTier   Membership = "free_tier"
)

type CampaignUser struct {
	Campaign       string     `json:"campaign" bson:"campaign"`
	LastName       string     `json:"lastname" bson:"lastname"`
	FirstName      string     `json:"firstname" bson:"firstname"`
	DeviceToken    string     `json:"device_token" bson:"device_token"`
	NativeLanguage string     `json:"native_language" bson:"native_language"`
	Membership     Membership `json:"membership" bson:"membership"`
	SortValue      int8       `json:"sort_value" bson:"sort_value"`
	VerifiedAt     time.Time  `json:"verified_at" bson:"verified_at"`
}
