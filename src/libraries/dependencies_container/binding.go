package dependencies_container

import "context"

type bindingType string

const (
	bindTransient bindingType = "transient"
	bindSingleton bindingType = "singleton"
)

type binding struct {
	bindingType
	closure func(context.Context) any
}
