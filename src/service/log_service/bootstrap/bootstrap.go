package bootstrap

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	config "duolingo/lib/config_reader"
	jr "duolingo/lib/config_reader/driver/reader/json"
	lb "duolingo/lib/log/grpc_service"
	sv "duolingo/lib/service_container"
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
	bindLogServer()
}

func bindContext() {
	container.BindSingleton("server.ctx", func() any { return ctx })
	container.BindSingleton("server.ctx_cancel", func() any { return cancel })
}

func bindConfigReader() {
	container.BindSingleton("config", func() any {
		conf := jr.NewJsonReader(filepath.Join(".", "config"))
		return conf
	})
}

func bindLogServer() {
	conf := container.Resolve("config").(config.ConfigReader)
	logSrvAddr := fmt.Sprintf("0.0.0.0:%v", conf.Get("log_service.server.port", "8002"))
	dbName := conf.Get("db.campaign.name", "")
	mongoUri := fmt.Sprintf("mongodb://%v:%v@%v:%v/",
		url.QueryEscape(conf.Get("db.campaign.user", "")),
		url.QueryEscape(conf.Get("db.campaign.password", "")),
		conf.Get("db.campaign.host", ""),
		conf.Get("db.campaign.port", ""),
	)
	container.BindSingleton("server.grpc_logger_server", func() any {
		srv, err := lb.
			NewLoggerGRPCServerBuilder(ctx, logSrvAddr).
			UseMongoLoggerService(mongoUri, dbName, "log").
			GetServer()
		if err != nil {
			panic(err)
		}
		return srv
	})
}
