package codebase

import (
	"errors"
	"fmt"
	"time"
)

type ErrPostpone time.Duration

func (e ErrPostpone) Error() string {
	return fmt.Sprintf("postpone for: %s", time.Duration(e).String())
}

func (e ErrPostpone) D() time.Duration {
	return time.Duration(e)
}

func IsErrPostpone(err error) bool {
	d := ErrPostpone(time.Second)
	return errors.As(err, &d)
}
