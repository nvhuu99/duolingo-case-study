package trace

import (
	"fmt"
)

type DataBag map[string]any

func NewDataBag(params ...any) DataBag {
	if len(params)%2 != 0 {
		panic(fmt.Sprintf(
			"NewDataBag: expected even number of arguments (key-value pairs), got %d",
			len(params),
		))
	}

	data := make(map[string]any)

	for i := 0; i < len(params); i += 2 {
		key, ok := params[i].(string)
		if !ok {
			panic(fmt.Sprintf(
				"NewDataBag: key at index %d is not a string (type %T)",
				i, params[i],
			))
		}
		data[key] = params[i+1]
	}

	return data
}

func (data DataBag) Merge(target DataBag) DataBag {
	merge := make(map[string]any)
	for k := range data {
		merge[k] = data[k]
	}
	for k := range target {
		merge[k] = target[k]
	}
	return merge
}

func (data DataBag) Exists(key string) bool {
	_, exists := data[key]
	return exists
}

func (data DataBag) Get(key string) string {
	if !data.Exists(key) {
		return ""
	}
	return fmt.Sprintf("%v", data[key])
}

func (data DataBag) GetAny(key string) any {
	if !data.Exists(key) {
		return ""
	}
	return data[key]
}


