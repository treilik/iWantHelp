package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type graphWriter interface {
	walder.NodeCreater
	walder.EdgeCreater
}

// Reachable transfers a subgraph from 'from' to 'to'
// if multiple nodes have the same string representation the transfered graph might not be the total reachable subgraph
func Reachable(start fmt.Stringer, from walder.GraphDirected, to graphWriter) error {
	if start == nil {
		return fmt.Errorf("start-node is nil")
	}
	if from == nil {
		return fmt.Errorf("from-graph is nil")
	}
	if to == nil {
		return fmt.Errorf("to-graph is nil")
	}

	seen := make(map[string]fmt.Stringer) // newNode lookup and cycle prevention
	s := start.String()
	n, err := to.NodeCreate(start)
	seen[s] = n
	if err != nil {
		return err
	}
	return reachable(seen, start, from, to)
}

func reachable(seen map[string]fmt.Stringer, cur fmt.Stringer, from walder.GraphDirected, to graphWriter) error {
	out, err := from.Outgoing(cur)
	if err != nil {
		return err
	}
	newStart, ok := seen[cur.String()] // This assums stable String()
	if !ok {
		return fmt.Errorf("node cant be identified by string: %#v", cur)
	}
	for _, o := range out {
		s := o.String()
		if n, in := seen[s]; in {
			err := to.EdgeCreate(newStart, n) // TODO if the string of the node is not unique, this might omit a reachable subgraph
			if err != nil {
				return err
			}
			continue
		}
		n, err := to.NodeCreate(o)
		if err != nil {
			return err
		}
		seen[s] = n
		newEnd := seen[o.String()]
		err = to.EdgeCreate(newStart, newEnd)
		if err != nil {
			return err
		}
		err = reachable(seen, o, from, to)
		if err != nil {
			return err
		}
	}
	return nil
}
