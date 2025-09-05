package message

type Platform string

type Priority string

type Visibility string

const (
	IOS     Platform = "ios"
	Android Platform = "android"
)

const (
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"

	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
	VisibilitySecret  Visibility = "secret"
)

func Platforms(p ...string) []Platform {
	result := make([]Platform, len(p))
	for i := range p {
		result = append(result, Platform(p[i]))
	}
	return result
}

func StrPlatforms(p ...Platform) []string {
	result := make([]string, len(p))
	for i := range p {
		result = append(result, string(p[i]))
	}
	return result
}
