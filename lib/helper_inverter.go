package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type inverter struct {
	invert walder.GraphDirected
}

var _ walder.GraphDirected = inverter{}

func (i inverter) String() string {
	if i.invert == nil {
		return "nothing to invert"
	}
	return fmt.Sprintf("inverting %s", i.invert.String())
}
func (i inverter) HomeNodes() ([]fmt.Stringer, error) {
	return i.invert.HomeNodes()
}
func (i inverter) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	return i.invert.Outgoing(node)
}
func (i inverter) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	return i.invert.Incoming(node)
}
