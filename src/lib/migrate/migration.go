package migrate

type Migration struct {
	Version     string
	Name        string
	Body        []byte
}
