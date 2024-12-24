package migrate

import "fmt"

type Migration struct {
	Id          string
	Name        string
	BatchNumber string
	Status      MigrateStatus
	Body        []byte
}

func (m *Migration) StatusLog(migrType MigrateType) string {
	if m.Name == "" || m.Status == "" {
		return ""
	}
	prefix := ""
	if migrType == MigrateRollback {
		prefix = "rollback: "
	}
	return fmt.Sprintf("%v%-60s %v", prefix, m.Name, m.Status)
}
