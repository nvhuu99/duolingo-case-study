package event

type Subcriber interface {
	SubcriberId() string
	Notified(string, any)
}
