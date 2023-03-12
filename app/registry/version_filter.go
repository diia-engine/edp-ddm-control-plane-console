package registry

import (
	"ddm-admin-console/service/codebase"
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
)

type versionFilter struct {
	compareFunc func(o *version.Version) bool
}

func makeVersionFilter(pattern string) (*versionFilter, error) {
	if pattern == "" {
		return &versionFilter{}, nil
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

	return &versionFilter{
		compareFunc: compareFunc,
	}, nil
}

func compareFunctions(v *version.Version) map[string]func(o *version.Version) bool {
	return map[string]func(o *version.Version) bool{
		"==": v.Equal,
		"<=": v.GreaterThanOrEqual,
		">=": v.LessThanOrEqual,
		"<":  v.GreaterThan,
		">":  v.LessThan,
	}
}

func (vf *versionFilter) filterCodebases(in []codebase.Codebase) ([]codebase.Codebase, error) {
	if vf.compareFunc == nil {
		return in, nil
	}

	res := make([]codebase.Codebase, 0, len(in))

	for _, cb := range in {
		if cb.Version == nil {
			continue
		}

		if vf.compareFunc(cb.Version) {
			res = append(res, cb)
		}
	}

	return res, nil
}
