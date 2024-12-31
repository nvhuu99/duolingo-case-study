package main

import (
	"context"
	"flag"
	"log"

	"duolingo/lib/helper-functions"
	"duolingo/lib/migrate"
	mongo "duolingo/lib/migrate/driver/database"
	local "duolingo/lib/migrate/driver/source"
	container "duolingo/lib/service-container"
)

const (
	usage = `Usage: go run migrate.go <up|rollback> --db="<mongo|mysql>" --db-name="" --host="" --port="" --user="" --pwd="" --src="" --src-uri=""`
)

var (
	ctx context.Context
	cancel context.CancelFunc
)

func bind() {
	container.Bind("source.local", func() any {
		return local.New(ctx, cancel)
	})

	container.Bind("driver.mongo", func() any {
		return mongo.New(ctx)
	})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bind()

	var migrType migrate.MigrateType

	// Define flags for database connection configuration
	database := flag.String("db", "", "Database driver (e.g., 'mongo' for MongoDB, 'mysql' for MySQL)")
	dbName := flag.String("db-name", "", "Name of the database to connect to")
	host := flag.String("host", "", "Host address of the database (e.g., 'localhost')")
	port := flag.String("port", "", "Port number on which the database is running")
	usr := flag.String("user", "", "Username for authenticating with the database")
	pwd := flag.String("pwd", "", "Password for authenticating with the database")

	// Flags for migration source configuration
	src := flag.String("src", "", "Migration source driver (e.g., 'local' for LocalFile)")
	srcURI := flag.String("src-uri", "", "URI or path to the location of migration files")

	// Parse flags from the command-line
	flag.Parse()

	// Check if all required flags are provided, exit if missing
	if *database == "" || *dbName == "" || *host == "" || *port == "" || *usr == "" || *pwd == "" || *src == "" || *srcURI == "" {
		log.Panic("Error: Missing required flags.\n" + usage)
	}

	// Validate the database type
	allowedDatabases := []string{"mongo", "mysql"}
	if !helper.InArray(*database, allowedDatabases) {
		log.Panic("Error: Unsupported database type. Only 'mongo' or 'mysql' are supported.")
	}

	// Validate the migration source
	allowedSources := []string{"local"}
	if !helper.InArray(*src, allowedSources) {
		log.Panic("Error: Unsupported migration source. Only 'local' is supported.")
	}

	// Determine migration type (up or rollback)
	switch flag.Arg(0) {
	case "up":
		migrType = migrate.MigrateUp
	case "rollback":
		migrType = migrate.MigrateRollback
	default:
		log.Panic("Error: Invalid migration type. Only 'up' or 'rollback' are allowed.")
	}

	// Set up the migration source
	source := container.Resolve("source." + *src).(migrate.Source)
	if err := source.UseUri(*srcURI); err != nil {
		log.Panic("Error: Failed to use source URI:", err)
	}

	// Set up the database driver
	driver := container.Resolve("driver." + *database).(migrate.Database)
	driver.SetDatabase(*dbName)
	driver.SetConnection(*host, *port, *usr, *pwd)

	// Set up the migration object
	migr := migrate.New(ctx, cancel)
	migr.SetMigrationType(migrType)
	migr.SetDatabaseDriver(driver)
	migr.SetMigrationSource(source)

	// Start the migration process
	migr.Start()
}
