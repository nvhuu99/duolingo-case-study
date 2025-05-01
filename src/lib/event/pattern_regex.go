package event

import (
	"regexp"
)

type regexPattern struct {
	raw   string
	regex *regexp.Regexp
}

func newRegexPattern(r string) (*regexPattern, error) {
	compiled, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}
	return &regexPattern{raw: r, regex: compiled}, nil
}

func (p *regexPattern) isEmptyPattern() bool {
	return p.raw == ""
}

func (p *regexPattern) match(target string) bool {
	return p.regex.MatchString(target)
}

func (p *regexPattern) equal(target *regexPattern) bool {
	return p.raw == target.raw
}
