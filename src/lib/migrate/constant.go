package migrate

import "fmt"

type MigrateType string

const (
	NilVersion          string        = "-1"
	MigrateUp           MigrateType   = "up"
	MigrateRollback     MigrateType   = "rollback"
)

var (
	ErrEmptyMigration = "Source: nothing to migrate"
	ErrNextMigration  = "Source: can not load next migration for current version. You should call HasNext() before invoke Next()"
	ErrPreviousMigration  = "Source: can not load previous migration for current version. You should call HasPrev() before invoke Prev()"
	ErrFileNameFormat = func(name string) string {
		return fmt.Sprintf(
			"Source: migration file format error. Filename: %v. Example of a valid name 00001_create_user_table.json",
			name,
		)
	}
	ErrOperationTimeOut = func(op string) string {
		return fmt.Sprintf(
			"Migrate: operation did not complete before timeout. Operation: %v",
			op,
		)
	}
)

