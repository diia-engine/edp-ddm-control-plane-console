package codebase

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type ErrPostpone time.Duration

func (e ErrPostpone) Error() string {
	return fmt.Sprintf("postpone for: %s", time.Duration(e).String())
}

func IsErrPostpone(err error) bool {
	_, ok := errors.Cause(err).(ErrPostpone)
	return ok
}
