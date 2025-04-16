package event

import (
	"regexp"
)

type RegexPattern struct {
	raw   string
	regex *regexp.Regexp
}

func NewRegexPattern(r string) (*RegexPattern, error) {
	compiled, err := regexp.Compile(r)
	if err != nil {
		return nil, err
	}
	return &RegexPattern{raw: r, regex: compiled}, nil
}

func (p *RegexPattern) IsEmptyPattern() bool {
	return p.raw == ""
}

func (p *RegexPattern) Match(target string) bool {
	return p.regex.MatchString(target)
}

func (p *RegexPattern) Equal(target *RegexPattern) bool {
	return p.raw == target.raw
}
