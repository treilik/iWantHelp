package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

const (
	dimNode   = nodeType("dims")
	graphNode = nodeType("graph")
	rNode     = nodeType("root")
)

type navigator struct {
	walder *Walder
}

type nodeType string

func (n nodeType) String() string {
	return string(n)
}

var _ walder.Graph = navigator{}

func (n navigator) String() string {
	return "navigator"
}

func (n navigator) HomeNodes() ([]fmt.Stringer, error) {
	if n.walder == nil {
		return nil, fmt.Errorf("this is a bug")
	}
	return []fmt.Stringer{
		dimNode,
		graphNode,
	}, nil
}

var _ internal = &navigator{}

func (n *navigator) renew(w *Walder) {
	n.walder = w
	return
}

var _ walder.GraphDirected = navigator{}

func (n navigator) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	switch v := node.(type) {
	case nodeType:
		if v != rNode {
			return []fmt.Stringer{rNode}, nil
		}
		return []fmt.Stringer{}, nil
	case walder.Dimensioner:
		return []fmt.Stringer{dimNode}, nil
	case walder.Graph:
		return []fmt.Stringer{graphNode}, nil
	}
	return nil, fmt.Errorf("unhandled %#v", node)
}

func (n navigator) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	switch v := node.(type) {
	case nodeType:
		switch v {
		case rNode:
			return []fmt.Stringer{
				dimNode,
				graphNode,
			}, nil
		case dimNode:
			stringers := make([]fmt.Stringer, 0, len(n.walder.dims))
			for _, d := range n.walder.dims {
				stringers = append(stringers, d)
			}
			return stringers, nil
		case graphNode:
			stringers := make([]fmt.Stringer, 0, len(n.walder.graphs))
			stringers = append(stringers, n.walder.bindings)
			for _, g := range n.walder.graphs {
				stringers = append(stringers, g)
			}

			return stringers, nil
		}
	case walder.Dimensioner:
		return []fmt.Stringer{}, nil
	case walder.Graph:
		return []fmt.Stringer{}, nil
	}
	return nil, fmt.Errorf("unhandled %#v", node)
}
