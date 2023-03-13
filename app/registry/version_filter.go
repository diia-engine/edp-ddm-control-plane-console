package registry

import (
	"ddm-admin-console/service/codebase"
	"errors"
	"fmt"
	"regexp"

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
	return map[string]func(o *version.Version) bool{
		"==": v.Equal,
		"<=": v.GreaterThanOrEqual,
		">=": v.LessThanOrEqual,
		"<":  v.GreaterThan,
		">":  v.LessThan,
	}
}

func (vf *VersionFilter) CodebaseIsVersion(cb *codebase.Codebase) bool {
	if vf.compareFunc == nil {
		return true
	}

	if cb.Version == nil {
		return false
	}

	return vf.compareFunc(cb.Version)
}

func (vf *VersionFilter) FilterCodebases(in []codebase.Codebase) []codebase.Codebase {
	if vf.compareFunc == nil {
		return in
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

	return res
}
