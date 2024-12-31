package container

var (
	svContainer *ServiceContainer
)

type binding struct {
	singleton bool
	closure func() any
}

type ServiceContainer struct {
	bindings map[string]binding
	instances map[string]any
}

func Container() *ServiceContainer {
	if svContainer == nil {
		svContainer = &ServiceContainer{}
		svContainer.bindings = make(map[string]binding)
		svContainer.instances = make(map[string]any)
	}

	return svContainer
}

func Bind(name string, closure func() any) {
	container := Container()
	container.bindings[name] = binding{
		singleton: false,
		closure: closure,
	}
}

func BindSingleton(name string, closure func() any) {
	container := Container()
	container.bindings[name] = binding{
		singleton: true,
		closure: closure,
	}
}

func Resolve(name string) any {
	container := Container()

	binding, found := container.bindings[name]
	if !found {
		return nil
	}

	var inst any
	if !binding.singleton {
		inst = binding.closure()
	} else {
		if _, found := container.instances[name]; !found {
			container.instances[name] = binding.closure()
		}
		inst = container.instances[name]
	}

	return inst
}