package trace

import (
	"regexp"
)

var placeholderRegex = regexp.MustCompile(`<([^<>]+)>`)
var specials = `()[]{}-`
var sanitizeRegex = regexp.MustCompile(`([` + regexp.QuoteMeta(specials) + `])`)

type SpanNameTemplate string

func (template SpanNameTemplate) Matches(instance string) bool {
	sanitizedTemplate := sanitizeRegex.ReplaceAllString(string(template), `\$1`)
	templateRegex, err := regexp.Compile(
		placeholderRegex.ReplaceAllString(sanitizedTemplate, `(.+)`),
	)
	if err != nil {
		return false
	}
	return templateRegex.MatchString(instance)
}

func (template SpanNameTemplate) ExtractVariables(instance string) DataBag {
	// extract placeholder names from template
	placeholders := placeholderRegex.FindAllStringSubmatch(string(template), -1)
	if len(placeholders) == 0 {
		return nil
	}
	// build regex to extract variables
	sanitizedTemplate := sanitizeRegex.ReplaceAllString(string(template), `\$1`)
	valueRegex, err := regexp.Compile(
		placeholderRegex.ReplaceAllString(sanitizedTemplate, `(.+)`),
	)
	if err != nil {
		return nil
	}
	// extract values
	valueMatches := valueRegex.FindAllStringSubmatch(instance, -1)
	if len(valueMatches) == 0 {
		return nil
	}
	values := valueMatches[0][1:]
	// build result
	result := make(map[string]any)
	for i := range placeholders {
		if i < len(values) {
			result[placeholders[i][1]] = values[i]
		}
	}
	return result
}
