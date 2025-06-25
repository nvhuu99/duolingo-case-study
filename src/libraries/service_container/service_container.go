package service_container

import (
	ctx "context"
	"errors"
	"reflect"
	"strings"
	"sync"
)

var (
	container                *ServiceContainer
	ensureContainerSingleton sync.Once
)

type ServiceContainer struct {
	ctx ctx.Context

	instancesMu sync.Mutex
	instances   map[string]any

	bindingsMu sync.Mutex
	bindings   map[string]*binding
}

func Init(ctx ctx.Context) {
	ensureContainerSingleton.Do(func() {
		container = &ServiceContainer{
			ctx:       ctx,
			instances: make(map[string]any),
			bindings:  make(map[string]*binding),
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
	alias := getTypeAlias[Abstract]()
	BindAlias(alias, closure)
}

func BindSingleton[Abstract any](closure func(ctx ctx.Context) any) {
	alias := getTypeAlias[Abstract]()
	BindSingletonAlias(alias, closure)
}

func Resolve[Abstract any]() (Abstract, bool) {
	alias := getTypeAlias[Abstract]()
	return ResolveAlias[Abstract](alias)
}

func MustResolve[Abstract any]() Abstract {
	alias := getTypeAlias[Abstract]()
	return MustResolveAlias[Abstract](alias)
}

func BindAlias(alias string, closure func(ctx ctx.Context) any) {
	c := Container()
	c.bindingsMu.Lock()
	defer c.bindingsMu.Unlock()
	c.bindings[alias] = &binding{bindTransient, closure}
}

func BindSingletonAlias(alias string, closure func(ctx ctx.Context) any) {
	c := Container()
	c.bindingsMu.Lock()
	defer c.bindingsMu.Unlock()
	c.bindings[alias] = &binding{bindSingleton, closure}
}

func ResolveAlias[Abstract any](alias string) (Abstract, bool) {
	c := Container()

	var zero Abstract

	c.bindingsMu.Lock()
	binding, bound := c.bindings[alias]
	c.bindingsMu.Unlock()

	if !bound {
		return zero, false
	}

	var instance any
	if binding.bindingType == bindSingleton {
		c.instancesMu.Lock()
		instance = c.instances[alias]
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
		c.instances[alias] = casted
		c.instancesMu.Unlock()
	}

	return casted, true
}

func MustResolveAlias[Abstract any](alias string) Abstract {
	instance, ok := ResolveAlias[Abstract](alias)
	if !ok {
		panic(errors.New("fail to resolve: " + alias))
	}
	return instance
}

func getTypeAlias[Abstract any]() string {
	typ := reflect.TypeOf((*Abstract)(nil)).Elem()
	return strings.Join([]string{typ.PkgPath(), typ.String()}, "/")
}
