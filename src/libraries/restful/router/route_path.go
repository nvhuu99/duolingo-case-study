package router

import (
	"net/url"
	"strings"
)

func cleanPathArr(path string) []string {
	parts := []string{}
	for _, p := range strings.Split(path, "/") {
		if len(p) > 0 {
			if escaped, err := url.PathUnescape(p); err == nil {
				parts = append(parts, escaped)
			} else {
				parts = append(parts, p)
			}
		}
	}

	return parts
}

func extractPathArgs(pattern string, requestPath string) map[string]string {
	paths := cleanPathArr(requestPath)
	patterns := cleanPathArr(pattern)
	if len(paths) != len(patterns) {
		return make(map[string]string)
	}

	result := make(map[string]string)
	for i := range patterns {
		if paths[i] == patterns[i] {
			continue
		}
		if isPathArg(patterns[i]) {
			key := strings.Trim(patterns[i], "{}")
			result[key] = paths[i]
		} else {
			return make(map[string]string)
		}
	}
	return result
}

func isPathArg(val string) bool {
	return strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}")
}
