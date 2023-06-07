package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type changer struct {
	dim  walder.Dimensions
	root origin
}

type origin fmt.Stringer

func (c changer) String() string {
	return "changer"
}

func (c changer) HomeNodes() ([]fmt.Stringer, error) {
	dims, err := c.dim.Dimensions(c.root)
	if err != nil {
		return nil, err
	}
	stringerList := make([]fmt.Stringer, 0, len(dims))
	for _, d := range dims {
		stringerList = append(stringerList, d)
	}
	return stringerList, nil
}

func (c changer) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	if _, ok := node.(origin); ok {
		return c.HomeNodes()
	}
	return []fmt.Stringer{}, nil
}

func (c changer) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	if _, ok := node.(origin); ok {
		return []fmt.Stringer{}, nil
	}
	return []fmt.Stringer{c.root}, nil
}
