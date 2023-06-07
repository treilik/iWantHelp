package walder

import (
	"fmt"
	"io"
)

// Dimensioner is a generator for Graphs in the Dimension
type Dimensioner interface {
	fmt.Stringer
	New() (Graph, error)
}

// Dimensions is a graph which can tell if a node is part of more graphs (has other Dimensions).
type Dimensions interface {
	fmt.Stringer
	Dimensions(fmt.Stringer) ([]Graph, error)
}

// OpenReader is a interface for graph adapters which can be opened from a io.Reader.
type OpenReader interface {
	fmt.Stringer
	Open(from io.Reader) (Graph, error)
}

// NodeOpener is a interface for a graph adapter which can be opened from fmt.Stringers (nodes).
type NodeOpener interface {
	fmt.Stringer
	NodeOpen(...fmt.Stringer) (Graph, error)
}

// Wraper is a interface to create a graph which wraps around an other graph and thus can operate on the later
type Wraper interface {
	fmt.Stringer
	Wrap(Graph) (Graph, error)
}
