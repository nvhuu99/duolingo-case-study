package main

import (
	"context"
	config "duolingo/lib/config_reader"
	"duolingo/lib/migrate"
	"duolingo/lib/migrate/driver/database/mongo"
	"duolingo/lib/migrate/driver/source/local"
	"flag"
	"log"
	"path/filepath"
	"runtime"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, caller, _, _ := runtime.Caller(0)
	dir := filepath.Dir(caller)
	config := config.NewJsonReader(filepath.Join(dir, "..", "..", "..", "config"))

	flag.Parse()

	var migrType migrate.MigrateType
	flg := flag.Arg(0)
	switch flg {
	case "", "up":
		migrType = migrate.MigrateUp
	case "rollback":
		migrType = migrate.MigrateRollback
	default:
		log.Fatal("Error: Invalid migration type. Only 'up' or 'rollback' are allowed.")
	}

	// Set up the migration source
	srcUri := filepath.Join(dir, "..", "migration")
	source := local.New(ctx, cancel)
	if err := source.UseUri(srcUri); err != nil {
		log.Fatal("Error: Failed to use source URI:", err)
	}
	go func() {
		<-ctx.Done()
		if source.HasError() {
			log.Fatal(source.Error())
		}
	}()

	// Set up the database driver
	driver := mongo.New(ctx)
	driver.SetDatabase(config.Get("db.campaign.name", ""))
	driver.SetConnection(
		config.Get("db.campaign.host", ""),
		config.Get("db.campaign.port", ""),
		config.Get("db.campaign.user", ""),
		config.Get("db.campaign.password", ""),
	)

	// Set up the migration object
	migr := migrate.New(ctx, cancel)
	migr.SetMigrationType(migrType)
	migr.SetDatabaseDriver(driver)
	migr.SetMigrationSource(source)

	// Start the migration process
	migr.Start()
}
