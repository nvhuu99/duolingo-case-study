package container

import (
	"context"
	"duolingo/lib/config"
	"duolingo/lib/migrate"
	mongo "duolingo/lib/migrate/driver/database"
	local "duolingo/lib/migrate/driver/source"
)

var (
	sourceContainer = make(map[string]migrate.Source)
	databaseContainer = make(map[string]migrate.Database)
	readerContainer = make(map[string]config.ConfigReader)
)

func MakeSource(name string, ctx context.Context, cancel context.CancelFunc) migrate.Source {
	if _, exists := sourceContainer[name]; !exists {
		switch name {
		case "local":
			sourceContainer[name] = local.New(ctx, cancel)
		default:
			return nil
		}
	}
	return sourceContainer[name]
}

func MakeDatabasae(name string, ctx context.Context, cancel context.CancelFunc) migrate.Database {
	if _, exists := databaseContainer[name]; !exists {
		switch name {
		case "mongo":
			databaseContainer[name] = mongo.New(ctx)
		default:
			return nil
		}
	}
	return databaseContainer[name]
}

func MakeConfigReader(name string, rootDir string) config.ConfigReader {
	if _, exists := readerContainer[name]; !exists {
		switch name {
		case "json":
			readerContainer[name] = config.NewJsonReader(rootDir)
		default:
			return nil
		}
	}
	return readerContainer[name]
}