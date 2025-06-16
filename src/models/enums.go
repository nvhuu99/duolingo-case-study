package models

type Membership string

type NativeLanguage string

const (
	MembershipPremium      Membership = "premium"
	MembershipFreeTier     Membership = "free_tier"
	MembershipSubscription Membership = "subscription"

	LanguageEN NativeLanguage = "en"
	LanguageVN NativeLanguage = "vn"
	LanguageJP NativeLanguage = "JP"
)
