package models

type Membership string

type NativeLanguage string

const (
	MEMBERSHIP_PREMIUM      Membership = "premium"
	MEMBERSHIP_FREE_TIER    Membership = "free_tier"
	MEMBERSHIP_SUBSCRIPTION Membership = "subscription"

	LANGUAGE_EN NativeLanguage = "en"
	LANGUAGE_VN NativeLanguage = "vn"
	LANGUAGE_JP NativeLanguage = "JP"
)
