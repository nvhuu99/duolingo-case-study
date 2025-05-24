package rest_http

import (
	"errors"
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

func (p *Value) Bool() (bool, bool) {
	asBool, ok := p.value.(bool)
	if !ok {
		return false, false
	}
	return asBool, true
}

func (p *Value) Int64() (int64, bool) {
	strVal := p.Str()
	intVal, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return 0, false
	}
	return intVal, true
}


func (p *Value) StrArr() ([]string, error) {
	asArr, ok := p.value.([]any)
	if !ok {
		return []string{}, errors.New("the input is not an array")
	}
	result := make([]string, len(asArr))
	for i, v := range asArr {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result, nil 
}