package bootstrap

import (
	"context"
	container "duolingo/libraries/service_container"
)

func Bootstrap() {
	container.Init(context.Background())
	BindConnections()
	BindPublisher()
}
