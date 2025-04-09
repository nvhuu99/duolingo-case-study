package rest_http

import (
	"fmt"
	"strconv"
)

type Value struct {
	value any
}

func (p *Value) Raw() any {
	return p.value
}

func (p *Value) Str() string {
	return fmt.Sprintf("%v", p.value)
}

func (p *Value) Int() (int, bool) {
	strVal := p.Str()
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, false
	}
	return intVal, true
}
