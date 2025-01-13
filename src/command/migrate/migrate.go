package main

import (
	"context"
	"flag"
	"log"

	"duolingo/lib/helper-functions"
	"duolingo/lib/migrate"
	"duolingo/lib/migrate/driver/database/mongo"
	"duolingo/lib/migrate/driver/database/mysql"
	"duolingo/lib/migrate/driver/source/local"
	sv "duolingo/lib/service-container"

)

const (
	usage = `Usage: go run migrate.go <up|rollback> --db="<mongo|mysql>" --db-name="" --host="" --port="" --user="" --pwd="" --src="" --src-uri=""`
)

var (
	ctx    context.Context
	cancel context.CancelFunc
	container = sv.NewContainer()
)

func bind() {
	container.Bind("source.local", func() any {
		return local.New(ctx, cancel)
	})

	container.Bind("driver.mongo", func() any {
		return mongo.New(ctx)
	})

	container.Bind("driver.mysql", func() any {
		return mysql.New(ctx)
	})
}

func main() {
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	bind()

	// Define flags for database connection configuration
	database := flag.String("db", "", "Database driver (e.g., 'mongo' for MongoDB, 'mysql' for MySQL)")
	dbName := flag.String("db-name", "", "Name of the database to connect to")
	host := flag.String("host", "", "Host address of the database (e.g., 'localhost')")
	port := flag.String("port", "", "Port number on which the database is running")
	usr := flag.String("user", "", "Username for authenticating with the database")
	pwd := flag.String("pwd", "", "Password for authenticating with the database")

	// Flags for migration source configuration
	src := flag.String("src", "local", "Migration source driver (e.g., 'local' for LocalFile)")
	srcURI := flag.String("src-uri", "", "URI or path to the location of migration files")

	// Parse flags from the command-line
	flag.Parse()

	// Check if all required flags are provided, exit if missing
	if *database == "" || *dbName == "" || *host == "" || *port == "" || *usr == "" || *pwd == "" || *src == "" || *srcURI == "" {
		log.Fatal("Error: Missing required flags.\n" + usage)
	}

	// Validate the database type
	allowedDatabases := []string{"mongo", "mysql"}
	if !helper.InArray(*database, allowedDatabases) {
		log.Fatal("Error: Unsupported database type. Only 'mongo' or 'mysql' are supported.")
	}

	// Validate the migration source
	allowedSources := []string{"local"}
	if !helper.InArray(*src, allowedSources) {
		log.Fatal("Error: Unsupported migration source. Only 'local' is supported.")
	}

	// Determine migration type (up or rollback)
	var migrType migrate.MigrateType
	flg := flag.Arg(0)
	switch flg {
	case "up":
		migrType = migrate.MigrateUp
	case "rollback":
		migrType = migrate.MigrateRollback
	default:
		log.Fatal("Error: Invalid migration type. Only 'up' or 'rollback' are allowed.")
	}

	// Set up the migration source
	source := container.Resolve("source." + *src).(migrate.Source)
	if err := source.UseUri(*srcURI); err != nil {
		log.Fatal("Error: Failed to use source URI:", err)
	}
	go func() {
		<-ctx.Done()
		if source.HasError() {
			log.Fatal(source.Error())
		}
	}()

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
