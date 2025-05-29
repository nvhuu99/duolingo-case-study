package bootstrap

import (
	"context"
	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
	rest "duolingo/lib/rest_http"
	sv "duolingo/lib/service_container"
	"duolingo/service/metric_api/middleware"

	// "fmt"
	log_repo "duolingo/repository/log"
	"path/filepath"
)

var (
	container *sv.ServiceContainer
	ctx       context.Context
	cancel    context.CancelFunc
)

func Run() {
	container = sv.GetContainer()
	ctx, cancel = context.WithCancel(context.Background())

	bindContext()
	bindConfigReader()
	bindRestHttp()
	bindRepository()
}

func bindContext() {
	container.BindSingleton("server.ctx", func() any { return ctx })
	container.BindSingleton("server.ctx_cancel", func() any { return cancel })
}

func bindConfigReader() {
	container.BindSingleton("config", func() any {
		dir, _ := filepath.Abs(filepath.Join(".", "config"))
		conf := jr.NewJsonReader(dir)
		return conf
	})
}

func bindRestHttp() {
	// conf := container.Resolve("config").(config.ConfigReader)

	container.BindSingleton("rest.server", func() any {
		server := rest.
			NewServer("localhost:8003").
			WithMiddlewares("request", new(middleware.SetCORSPolicies))
			// NewServer(fmt.Sprintf("0.0.0.0:%v", conf.Get("input_message_api.server.port", "8003")))
			// WithMiddlewares("request", new(md.RequestBegin)).
			// WithMiddlewares("response", new(md.RequestEnd))

		return server
	})
}

func bindRepository() {
	conf := container.Resolve("config").(config.ConfigReader)
	container.BindSingleton("repo.log", func() any {
		repo := log_repo.NewLogRepo(ctx, conf.Get("db.campaign.name", ""))
		err := repo.SetConnection(
			// conf.Get("db.campaign.host", ""),
			"localhost",
			conf.Get("db.campaign.port", ""),
			conf.Get("db.campaign.user", ""),
			conf.Get("db.campaign.password", ""),
		)
		if err != nil {
			panic(err)
		}
		return repo
	})
}
