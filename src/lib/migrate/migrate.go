package migrate

import (
	"context"
	"duolingo/lib/helper_functions"
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	errMigrationFailure   = "Migrate: failed"
	upToDateMessage       = "Migrate: already up to date"
	errRollbackBatchEmpty = "Migrate: nothing to rollback"
	completeMessage       = "Migrate: completed"
	startMigrationMessage = "Migrate: running"
)

type Migrate struct {
	ctx      context.Context
	cancel   context.CancelFunc
	src      Source
	driver   Database
	migrType MigrateType
	err      error
}

func New(ctx context.Context, cancel context.CancelFunc) *Migrate {
	migr := Migrate{}
	migr.ctx = ctx
	migr.cancel = cancel

	return &migr
}

func (migr *Migrate) SetMigrationSource(src Source) *Migrate {
	migr.src = src
	return migr
}

func (migr *Migrate) SetDatabaseDriver(driver Database) *Migrate {
	migr.driver = driver
	return migr
}

func (migr *Migrate) SetMigrationType(mt MigrateType) *Migrate {
	migr.migrType = mt
	return migr
}

func (migr *Migrate) Start() {
	defer func() {
		if migr.err != nil {
			migr.cancel()
			log.Println(errMigrationFailure)
			log.Println(migr.err.Error())
		} else {
			migr.src.Close()
			log.Println(completeMessage)
		}
	}()

	// driver
	err := migr.driver.PrepareDatabase()
	if err != nil {
		migr.err = err
		return
	}
	batchNumber := migr.driver.BatchNumber()
	lastBatch, err := migr.driver.LastBatch()
	if err != nil {
		migr.err = err
		return
	}

	// build batch
	var batch []string
	if migr.migrType == MigrateRollback {
		batch = migr.makeBatchRollBack(lastBatch)
		if len(batch) == 0 {
			migr.err = errors.New(errRollbackBatchEmpty)
			return
		}
	} else {
		batch = migr.makeBatchMigrateUp(lastBatch)
		if len(batch) == 0 {
			log.Println(upToDateMessage)
			return
		}
	}

	migr.src.Open(batch)

	// run migration
	log.Println(startMigrationMessage)
	for {
		select {
		case <-migr.ctx.Done():
			if migr.src.HasError() {
				migr.err = err
			}
			return
		default:
			if !migr.run(batchNumber, lastBatch) {
				return
			}
		}
	}
}

func (migr *Migrate) makeBatchRollBack(lastBatch []Migration) []string {
	batch := make([]string, len(lastBatch))
	for i, m := range lastBatch {
		batch[i] = m.Name + ".rollback" + migr.driver.GetFileExt()
	}
	helper.ReverseSlice(batch)

	return batch
}

func (migr *Migrate) makeBatchMigrateUp(lastBatch []Migration) []string {
	var batch []string
	files := migr.src.List()
	helper.ReverseSlice(files)
	var lastMigration *Migration
	if len(lastBatch) > 0 {
		lastMigration = &lastBatch[len(lastBatch)-1]
	}
	for _, filename := range files {
		re, _ := regexp.Compile(`(.*)(\.[a-z]+$)`)
		parts := re.FindStringSubmatch(filename)
		if lastMigration != nil && parts[1] == lastMigration.Name {
			break
		}
		if strings.Contains(filename, ".rollback.") {
			continue
		}
		batch = append(batch, parts[1]+migr.driver.GetFileExt())
	}
	helper.ReverseSlice(batch)

	return batch
}

func (migr *Migrate) run(batchNumber int, lastBatch []Migration) bool {
	if !migr.src.HasNext() {
		return false
	}
	migration, err := migr.src.Next()
	if err != nil {
		migr.err = err
		return false
	}
	err = migr.driver.RunMigration(migration)
	if err != nil {
		migr.err = err
		migration.Status = MigrateFailed
	} else {
		migration.Status = MigrateFinished
	}

	if migr.migrType == MigrateUp {
		migration.BatchNumber = strconv.Itoa(batchNumber)
		migr.driver.SaveMigrationRecord(migration)
	} else {
		for _, last := range lastBatch {
			if last.Name == migration.Name {
				migration.Status = MigrateFinished
				migr.driver.DeleteMigrationRecord(&last)
			}
		}
	}

	log.Println(migration.StatusLog(migr.migrType))

	return migr.err == nil
}
