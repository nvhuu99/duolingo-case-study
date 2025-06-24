package service_container

import (
	ctx "context"
	"errors"
	"reflect"
	"sync"
)

var (
	container                *ServiceContainer
	ensureContainerSingleton sync.Once
)

type ServiceContainer struct {
	ctx ctx.Context

	instancesMu sync.Mutex
	instances   map[reflect.Type]any

	bindingsMu sync.Mutex
	bindings   map[reflect.Type]*binding
}

func Init(ctx ctx.Context) {
	ensureContainerSingleton.Do(func() {
		container = &ServiceContainer{
			ctx:       ctx,
			instances: make(map[reflect.Type]any),
			bindings:  make(map[reflect.Type]*binding),
		}
	})
}

func Container() *ServiceContainer {
	if container == nil {
		panic(errors.New("Init() must be called before using the container"))
	}
	return container
}

func Bind[Abstract any](closure func(ctx ctx.Context) any) {
	c := Container()
	c.bindingsMu.Lock()
	defer c.bindingsMu.Unlock()

	typ := reflect.TypeOf((*Abstract)(nil)).Elem()
	c.bindings[typ] = &binding{bindTransient, closure}
}

func BindSingleton[Abstract any](closure func(ctx ctx.Context) any) {
	c := Container()
	c.bindingsMu.Lock()
	defer c.bindingsMu.Unlock()

	typ := reflect.TypeOf((*Abstract)(nil)).Elem()
	c.bindings[typ] = &binding{bindSingleton, closure}
}

func MustResolve[Abstract any]() Abstract {
	instance, ok := Resolve[Abstract]()
	if !ok {
		typ := reflect.TypeOf((*Abstract)(nil)).Elem()
		panic(errors.New("fail to resolve: " + typ.String()))
	}
	return instance
}

func Resolve[Abstract any]() (Abstract, bool) {
	c := Container()

	var zero Abstract
	typ := reflect.TypeOf((*Abstract)(nil)).Elem()

	c.bindingsMu.Lock()
	binding, bound := c.bindings[typ]
	c.bindingsMu.Unlock()

	if !bound {
		return zero, false
	}

	var instance any
	if binding.bindingType == bindSingleton {
		c.instancesMu.Lock()
		instance = c.instances[typ]
		c.instancesMu.Unlock()
	}
	if instance == nil {
		instance = binding.closure(c.ctx)
	}

	casted, ok := instance.(Abstract)
	if !ok {
		return zero, false
	}

	if binding.bindingType == bindSingleton {
		c.instancesMu.Lock()
		c.instances[typ] = casted
		c.instancesMu.Unlock()
	}

	return casted, true
}
