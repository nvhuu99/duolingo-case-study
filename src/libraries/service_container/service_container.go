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

	mu        sync.Mutex
	instances map[reflect.Type]any
	bindings  map[reflect.Type]*binding
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
	c.mu.Lock()
	defer c.mu.Unlock()

	typ := reflect.TypeOf((*Abstract)(nil)).Elem()
	c.bindings[typ] = &binding{bindTransient, closure}
}

func BindSingleton[Abstract any](closure func(ctx ctx.Context) any) {
	c := Container()
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero Abstract
	typ := reflect.TypeOf((*Abstract)(nil)).Elem()

	binding, bound := c.bindings[typ]
	if !bound {
		return zero, false
	}

	var instance any
	if binding.bindingType == bindSingleton {
		if _, resolved := c.instances[typ]; resolved {
			instance = c.instances[typ]
		}
	}
	if instance == nil {
		instance = binding.closure(c.ctx)
	}

	casted, ok := instance.(Abstract)
	if !ok {
		return zero, false
	}

	if binding.bindingType == bindSingleton {
		c.instances[typ] = casted
	}

	return casted, true
}
