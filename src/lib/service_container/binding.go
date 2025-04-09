package container

type binding struct {
	singleton bool
	closure   func() any
}
