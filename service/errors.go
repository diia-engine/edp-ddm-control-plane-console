package service

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}

func IsErrNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}
