package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type topo struct {
	origin    walder.GraphDirected
	adjacency map[string]neighbors
}

var _ walder.Graph = topo{}

func (t topo) String() string {
	if t.origin == nil {
		return "" // TODO fix
	}
	return t.origin.String()
}
func (t topo) HomeNodes() ([]fmt.Stringer, error) {
	if t.origin == nil {
		return nil, fmt.Errorf("no graph to operate on")
	}
	return t.origin.HomeNodes()
}

type neighbors struct {
	node fmt.Stringer
	in   []fmt.Stringer
	out  []fmt.Stringer
}

func (t topo) TopoSort() ([]fmt.Stringer, error) {
	aller, ok := t.origin.(walder.NodeAller)
	if !ok {
		return nil, fmt.Errorf("%T does not implement %s", t.origin, nodeAllerString)
	}
	all, err := aller.NodeAll()
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return all, nil
	}

	t.adjacency = make(map[string]neighbors)
	var startNodes []fmt.Stringer

	for _, node := range all {
		in, err := t.origin.Incoming(node)
		if err != nil {
			return nil, err
		}
		if len(in) == 0 {
			startNodes = append(startNodes, node)
		}
		out, err := t.origin.Outgoing(node)
		if err != nil {
			return nil, err
		}
		key := node.String()
		_, ok := t.adjacency[key]
		if ok {
			return nil, fmt.Errorf("to nodes return the same string: %s", key)
		}
		if second := node.String(); key != second {
			return nil, fmt.Errorf("the same node has returned different strings and thus the string can't be use to identify the node: %s, %s", key, second)
		}
		t.adjacency[key] = neighbors{node: node, in: in, out: out}
	}

	if len(startNodes) == 0 {
		return nil, fmt.Errorf("no nodes without incoming edges, thus the graph contains cycles and can not be topological sorted")
	}
	sortedNodes := make([]fmt.Stringer, 0, len(t.adjacency))

	for len(startNodes) > 0 {
		lenght := len(startNodes)

		cur := startNodes[lenght-1]
		sortedNodes = append(sortedNodes, cur)
		startNodes = startNodes[:lenght-1]

		key := cur.String()
		nb, ok := t.adjacency[key]
		if !ok {
			return nil, fmt.Errorf("the node %v has returned a different string then before and thus can't be used to identifie the node", cur)
		}
		out := nb.out
		for len(out) > 0 {
			toKey := out[len(out)-1].String()

			// remove edge from graph becaus it was added allready to the sorted Nodes
			// remove outedge
			out = out[:len(out)-1]

			// remove inedge
			var found bool
			in := t.adjacency[toKey]
			for i, n := range in.in {
				if n.String() != key {
					continue
				}
				var rest []fmt.Stringer
				if len(in.in) > i {
					rest = in.in[i+1:]
				}
				in.in = append(in.in[:i], rest...)
				t.adjacency[toKey] = in

				found = true
				break
			}

			// if there is a outgoing edge for one node there has to be a incoming node for an other and if there is none:
			if !found {
				return nil, fmt.Errorf("a node has returned a different string then before and thus can't be used to identify the node")
			}

			// we created a new start node by removing the last inoming node
			if len(t.adjacency[toKey].in) == 0 {
				startNodes = append(startNodes, t.adjacency[toKey].node)
			}
		}
		t.adjacency[key] = nb
	}
	for _, v := range t.adjacency {
		if len(v.in) != 0 {
			return sortedNodes, fmt.Errorf("the graph contains zykles and thus can not be topological sorted")
		}
	}
	return sortedNodes, nil
}
