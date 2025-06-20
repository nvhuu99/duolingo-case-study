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
