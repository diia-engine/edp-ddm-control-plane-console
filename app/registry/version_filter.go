package registry

import (
	"ddm-admin-console/service/codebase"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

type VersionFilter struct {
	compareFunc func(o *version.Version) bool
}

func MakeVersionFilter(pattern string) (*VersionFilter, error) {
	if pattern == "" {
		return &VersionFilter{}, nil
	}

	regPattern := regexp.MustCompile(`(<=|>=|==|>|<)(.+)`)
	elements := regPattern.FindStringSubmatch(pattern)
	if len(elements) < 3 {
		return nil, errors.New("wrong input pattern")
	}

	v, err := version.NewVersion(elements[2])
	if err != nil {
		return nil, fmt.Errorf("wrong version pattern, %w", err)
	}

	compareFunc, ok := compareFunctions(v)[elements[1]]
	if !ok {
		return nil, errors.New("wrong compare pattern")
	}

	return &VersionFilter{
		compareFunc: compareFunc,
	}, nil
}

func compareFunctions(v *version.Version) map[string]func(o *version.Version) bool {
	versionString := v.String()

	equalFunc := func(other *version.Version) bool {
		otherString := other.String()

		return strings.HasPrefix(otherString, versionString)
	}

	greaterFunc := func(other *version.Version) bool {
		if equalFunc(other) {
			return false
		}

		return v.GreaterThan(other)
	}

	lessFunc := func(other *version.Version) bool {
		if equalFunc(other) {
			return false
		}

		return v.LessThan(other)
	}

	return map[string]func(o *version.Version) bool{
		"==": equalFunc,
		"<":  greaterFunc,
		">":  lessFunc,
	}
}

// CheckCodebase checks if codebase version matches the filter.
func (vf *VersionFilter) CheckCodebase(cb *codebase.Codebase) bool {
	if vf.compareFunc == nil {
		return true
	}

	if cb.Version == nil {
		return false
	}

	return vf.compareFunc(cb.Version)
}
