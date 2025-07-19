package fake

type Animal interface {
	MakeSound() string
}

type Dog struct {
	Id string
}

func (d *Dog) MakeSound() string {
	return "woof woof from " + d.Id
}

type Cat struct {
	Id string
}

func (c *Cat) MakeSound() string {
	return "meow meow from " + c.Id
}
