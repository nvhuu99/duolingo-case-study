package resthttp

import (
	"fmt"
	"strconv"
)

type RequestParam struct {
	value any
}

func (p *RequestParam) Raw() any {
	return p.value
}

func (p *RequestParam) Str() string { 
	return fmt.Sprintf("%v", p.value)
}

func (p *RequestParam) Int() (int, bool) { 
	strVal := p.Str()
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, false
	}
	return intVal, true
}
