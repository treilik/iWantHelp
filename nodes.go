package walder

func NewError(e error) Error {
	return Error{e}
}

type Error struct {
	err error
}

func (e Error) String() string {
	if e.err == nil {
		return "<nil> error"
	}
	return e.err.Error()
}

var _ error = Error{}

func (e Error) Error() string {
	return e.err.Error()
}

type Pather interface {
	Path() (string, error)
}
