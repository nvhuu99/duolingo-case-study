package migrate

type Migration struct {
	Id			string
	Version     string
	Name        string
	Status		MigrateStatus
	Body        []byte
}
