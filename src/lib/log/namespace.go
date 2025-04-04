package log

import "strings"

func Namespace(parts ...string) string {
	return strings.Join(parts, ":")
}
