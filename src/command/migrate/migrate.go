package main

import (
	"context"
	"flag"
	"log"
	"os"

	"duolingo/lib/helper-functions"
	"duolingo/lib/migrate"
	"duolingo/lib/migrate/container"
)

const (
	usage = `Usage: go run migrate.go <up|rollback> --db="<mongo|mysql>" --db-name="" --host="" --port="" --user="" --pwd="" --src="" --src-uri=""`
)

func main() {
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
		log.Fatalln("Error: Missing required flags.\n" + usage)
		os.Exit(1)
	}

	// Validate the database type
	allowedDatabases := []string{"mongo", "mysql"}
	if !helper.InArray(*database, allowedDatabases) {
		log.Fatalln("Error: Unsupported database type. Only 'mongo' or 'mysql' are supported.")
		os.Exit(1)
	}

	// Validate the migration source
	allowedSources := []string{"local"}
	if !helper.InArray(*src, allowedSources) {
		log.Fatalln("Error: Unsupported migration source. Only 'local' is supported.")
		os.Exit(1)
	}

	// Determine migration type (up or rollback)
	switch flag.Arg(0) {
	case "up":
		migrType = migrate.MigrateUp
	case "rollback":
		migrType = migrate.MigrateRollback
	default:
		log.Fatalln("Error: Invalid migration type. Only 'up' or 'rollback' are allowed.")
		os.Exit(1)
	}

	// Set up the migration context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up the migration source
	source := container.MakeSource(*src, ctx, cancel)
	if err := source.UseUri(*srcURI); err != nil {
		log.Fatalln("Error: Failed to use source URI:", err)
		os.Exit(1)
	}

	// Set up the database driver
	driver := container.MakeDatabasae(*database, ctx, cancel)
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
