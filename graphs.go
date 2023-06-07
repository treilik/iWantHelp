package walder

import (
	"fmt"
	"io"
)

// Graph is the interface which is expected from every graph adapter implementation.
type Graph interface {
	// every graph adapter should be able to represent it self via a string
	// this makes every graph a possible node in an other graph
	fmt.Stringer

	// every graph should be able to provide some of its nodes
	HomeNodes() ([]fmt.Stringer, error)
}

func GraphTest(g Graph) error {
	if g == nil {
		return fmt.Errorf("received nil value")
	}
	home, err := g.HomeNodes()
	if err != nil {
		return fmt.Errorf("received unexpected error: %w", err)
	}
	if home == nil {
		return fmt.Errorf("received nil, use empty slice instead")
	}

	// if len(home) == 0 // TODO warn?

	for _, h := range home {
		if h == nil {
			return fmt.Errorf("HomeNodes returned a nil value")
		}
		first, second := h.String(), h.String()
		if first != second {
			return fmt.Errorf("node %#v returned different string on repeated call: '%s' != '%s'", h, first, second)
		}
	}
	return nil
}

type GraphNeighbors interface {
	Graph
	Neighbors(fmt.Stringer) ([]fmt.Stringer, error)
}

func GraphNeighborsTest(g GraphNeighbors) error {
	if g == nil {
		return fmt.Errorf("received nil value")
	}
	home, err := g.HomeNodes()
	if err != nil {
		return err
	}
	if len(home) == 0 {
		return fmt.Errorf("cant test if no nodes are provided")
	}

	for _, h := range home {
		if h == nil {
			return fmt.Errorf("received nil value")
		}

		neighbors, err := g.Neighbors(h)
		if err != nil {
			return err
		}

		for _, n := range neighbors {
			if n == nil {
				return fmt.Errorf("received nil value")
			}
			if n.String() != h.String() {
				continue
			}
			return fmt.Errorf("inconsistency: node '%#v' is neighbor of node '%#v' but second is not neighbor of first", h, n)
		}
	}
	return nil
}

type GraphIncoming interface {
	Graph
	Incoming(fmt.Stringer) ([]fmt.Stringer, error)
}

type GraphOutgoing interface {
	Graph
	Outgoing(fmt.Stringer) ([]fmt.Stringer, error)
}

type GraphDirectedTree interface {
	Parent(fmt.Stringer) (fmt.Stringer, error)
	Children(fmt.Stringer) ([]fmt.Stringer, error)
}

// GraphDirected is a interface used to navigate directed graphs.
type GraphDirected interface {
	GraphIncoming
	GraphOutgoing
}
type GraphEqualer interface {
	Graph
	Equal(fmt.Stringer, fmt.Stringer) (bool, error)
}
type GraphLesser interface {
	Graph
	Less(fmt.Stringer, fmt.Stringer) (bool, error)
}

type GraphConstrainer interface {
	Graph
	Constrain() map[string]bool
}

func GraphConstrainerTest(g GraphConstrainer) error {
	if g == nil {
		return fmt.Errorf("received nil value")
	}
	for k, _ := range g.Constrain() {
		if _, ok := GraphConstrainAllowed[Constrain(k)]; !ok {
			return fmt.Errorf("unkown constrain: '%s' please use one of: %v", k, GraphConstrainAllowed)
		}
	}
	return nil
}

// GetReader is a interface for graph adapters which can provide a io.Reader which captures the state of the graph.
type GetReader interface {
	Graph
	GetReader() (io.Reader, error)
}

// NodeReader is a interface for graph adapters which can provide a io.Reader for there nodes.
type NodeReader interface {
	Graph
	NodeRead(fmt.Stringer) (io.Reader, error)
}

// NodeNamer is a interface for graph adapters which can rename there Nodes.
type NodeNamer interface {
	Graph
	NodeName(node fmt.Stringer, newName string) (fmt.Stringer, error)
	NodeUpdater
}

// NodeCreater is a interface for graph adapters which can create new Nodes.
type NodeCreater interface {
	Graph
	NodeCreate(input fmt.Stringer) (fmt.Stringer, error)
}

// NodeFromCreater is a interface to create nodes which need to have a connection from a nother node at creation
type NodeFromCreater interface {
	Graph
	NodeFromCreate(input fmt.Stringer, from ...fmt.Stringer) (fmt.Stringer, error)
}

// NodeToCreater is a interface to create nodes which need to have a connection to a nother node at creation
type NodeToCreater interface {
	Graph
	NodeToCreate(input fmt.Stringer, from ...fmt.Stringer) (fmt.Stringer, error)
}

// NodeUpdater is a interface for graph adpters which can modify there nodes.
type NodeUpdater interface {
	Graph
	NodeUpdate(fmt.Stringer) (fmt.Stringer, error)
}

// EdgeCreater is a interface for graph adapters which can create a esge.
type EdgeCreater interface {
	Graph
	NodeUpdater
	EdgeCreate(from, to fmt.Stringer) error
}

// NodeDeleter is a interface for graph adapters which can delete a node.
type NodeDeleter interface {
	Graph
	NodeUpdater
	NodeDelete(toDelete fmt.Stringer) error
}

// EdgeDeleter is a interface for graph adapters which can delete a edge.
type EdgeDeleter interface {
	Graph
	NodeUpdater
	EdgeDelete(from, to fmt.Stringer) error
}
type EdgeInverter interface {
	GraphDirected
	EdgeInvert(from, to fmt.Stringer) error
}

// NodeWriter is a interface for graph adapters which can provide a io.Writer for its nodes.
type NodeWriter interface {
	Graph
	NodeUpdater
	// io.Writer is excpected to have side effekts onto the graph
	// so its guaranteed that all changes until the Close call are preserved.
	NodeWrite(fmt.Stringer) (io.WriteCloser, error)
}

type NodeAller interface {
	Graph
	NodeAll() ([]fmt.Stringer, error)
}

type NodeSwaper interface {
	Graph
	NodeSwap(_, _ fmt.Stringer) error
}
type EdgeMover interface {
	Graph
	EdgeMove(toMove, from, to fmt.Stringer) error
}

type Meta interface {
	Graph
	Get() (Graph, error)
	Set(Graph) error
}

type Executor interface {
	Graph
	Execute(node fmt.Stringer) error
}

type Typer interface {
	GetType(node fmt.Stringer) (string, error)
}

type NodeTypedCreator interface {
	Graph
	GetTypes() ([]fmt.Stringer, error)
	NodeTypedCreate(Type fmt.Stringer, input fmt.Stringer) (fmt.Stringer, error)
}

type NodeLabeler interface {
	Graph
	NodeLabels(node fmt.Stringer) ([][2]string, error)
}
type NodeLabelAdder interface {
	NodeLabeler
	NodeLabelAdd(node fmt.Stringer, key, value string) (fmt.Stringer, error)
}
type EdgeLabeler interface {
	Graph
	EdgeLabels(from, to fmt.Stringer) ([][2]string, error)
}
type EdgeLabelAdder interface {
	EdgeLabeler
	EdgeLabelAdd(from, to fmt.Stringer, key, value string) error
}

type Closer interface {
	Close() error
}
type DimensionChanger interface {
	Graph
	DimensionGetAll() ([]fmt.Stringer, error)
	DimensionSet(dim fmt.Stringer) error
}

type GraphCreater interface {
	NodeCreater
	EdgeCreater
}
