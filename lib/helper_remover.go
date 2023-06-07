package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type dimension string

const (
	dimBool     dimension = "bool"
	dimFiltered dimension = "filtered"
)

type removerNode struct {
	name       string
	node       fmt.Stringer
	children   []removerNode
	removeFunc func(r *removerNode, node fmt.Stringer) (bool, error)
	origin     walder.GraphDirected
	id         int
}

func (r removerNode) String() string {
	if r.removeFunc == nil && r.node != nil {
		return r.node.String()
	}
	if r.name == outR {
		if r.node == nil {
			return outR
		}
		return fmt.Sprintf("out reachable of %s", r.node.String())
	}
	return fmt.Sprintf("%s of: %#v", r.name, r.children)
}

const (
	not = "NOT"
	and = "AND"
	or  = "OR"

	outR = "OUT"
	inR  = "IN"
)

var boolNodes = map[string]removerNode{
	not: {
		name: not,
		removeFunc: func(r *removerNode, node fmt.Stringer) (bool, error) {
			if len(r.children) != 1 {
				return false, fmt.Errorf("cant invert multiple nodes")
			}
			b, err := r.children[0].removeFunc(r, node)
			if err != nil {
				return !b, err
			}
			return !b, nil
		},
	},
	or: {
		name: or,
		removeFunc: func(r *removerNode, node fmt.Stringer) (bool, error) {
			for _, c := range r.children {
				b, err := c.removeFunc(r, node)
				if err != nil {
					return b, err
				}
				if b {
					return true, nil
				}
			}
			return false, nil
		},
	},
	and: {
		name: and,
		removeFunc: func(r *removerNode, node fmt.Stringer) (bool, error) {
			for _, c := range r.children {
				b, err := c.removeFunc(r, node)
				if err != nil {
					return b, err
				}
				if !b {
					return false, nil
				}
			}
			return true, nil
		},
	},
	outR: {
		name: outR,
		removeFunc: func(r *removerNode, node fmt.Stringer) (bool, error) {
			if len(r.children) != 0 {
				return false, fmt.Errorf("reachable is a leaf and should not have children but hast some: %#v", r.children)
			}
			return outReachable(r.origin, node, r.node)
		},
	},
	inR: {
		name: inR,
		removeFunc: func(r *removerNode, node fmt.Stringer) (bool, error) {
			if len(r.children) != 0 {
				return false, fmt.Errorf("reachable is a leaf and should not have children but hast some: %#v", r.children)
			}
			return inRechable(r.origin, node, r.node)
		},
	},
}

func (r removerNode) Remove(node fmt.Stringer) (bool, error) {
	if r.removeFunc == nil {
		return false, fmt.Errorf("no remove function set")
	}
	return r.removeFunc(&r, node)
}

type remover struct {
	startNodes []fmt.Stringer
	dim        dimension
	origin     walder.GraphDirected
	removeTree removerNode
	allNodes   map[int]removerNode
	lastID     int
}

var _ walder.DimensionChanger = &remover{}

func (r *remover) DimensionGetAll() ([]fmt.Stringer, error) {
	return []fmt.Stringer{
		stringer(dimBool),
		stringer(dimFiltered),
	}, nil
}
func (r *remover) DimensionSet(dim fmt.Stringer) error {
	switch dim := dim.String(); dim {
	case string(dimBool):
		r.dim = dimension(dim)
		return nil
	case string(dimFiltered):
		r.dim = dimension(dim)
		return nil
	}
	return fmt.Errorf("dimension '%s' not known to this graph", dim)
}

var _ walder.Graph = &remover{}

func (r *remover) String() string {
	if r.origin == nil {
		return "empty remover"
	}
	return fmt.Sprintf("remover of graph: %s", r.origin.String())
}
func (r *remover) HomeNodes() ([]fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("no underlying graph")
	}
	switch r.dim {
	case dimFiltered:
		return r.filteredHome()
	case dimBool:
		return r.boolHome()
	}
	return nil, fmt.Errorf("unhandled dimensions: %s", r.dim)
}
func (r *remover) filteredHome() ([]fmt.Stringer, error) {
	homes, err := r.origin.HomeNodes()
	if err != nil {
		return homes, err
	}
	return r.filterNode(homes...)
}
func (r *remover) boolHome() ([]fmt.Stringer, error) {
	var home []fmt.Stringer
	for _, v := range r.allNodes {
		home = append(home, v)
	}
	home = append(home, r.startNodes...)
	return home, nil
}
func (r *remover) filterNode(nodes ...fmt.Stringer) ([]fmt.Stringer, error) {
	if r.removeTree.removeFunc == nil {
		return nodes, nil
	}
	var filtered []fmt.Stringer
	for _, n := range nodes {
		r.removeTree.origin = r.origin
		remove, err := r.removeTree.Remove(n)
		if err != nil {
			return nil, err
		}
		if remove {
			continue
		}
		filtered = append(filtered, n)
	}
	return filtered, nil
}

var _ walder.GraphDirected = &remover{}

func (r *remover) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	switch r.dim {
	case dimFiltered:
		return r.filteredIncoming(node)
	case dimBool:
		return r.boolIncoming(node)
	}
	return nil, fmt.Errorf("unhandled dimensions: %s", r.dim)
}

func (r *remover) filteredIncoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("no underlying graph")
	}
	_, err := r.filterNode(node)
	if err != nil {
		return nil, err
	}
	inR, err := r.origin.Incoming(node)
	if err != nil {
		return nil, err
	}
	return r.filterNode(inR...)
}

func (r *remover) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	switch r.dim {
	case dimFiltered:
		return r.filteredOutgoing(node)
	case dimBool:
		return r.boolOutgoing(node)
	}
	return nil, fmt.Errorf("unhandled dimensions: %s", r.dim)
}

func (r *remover) filteredOutgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("no underlying graph")
	}
	_, err := r.filterNode(node)
	if err != nil {
		return nil, err
	}
	outR, err := r.origin.Outgoing(node)
	if err != nil {
		return nil, err
	}
	return r.filterNode(outR...)
}
func (r *remover) boolOutgoing(n fmt.Stringer) ([]fmt.Stringer, error) {
	node, ok := n.(removerNode)
	if !ok {
		return nil, nil
	}
	switch node.name {
	case not:
		return []fmt.Stringer{node.children[0]}, nil
	case and:
		var children []fmt.Stringer
		for _, c := range node.children {
			children = append(children, c)
		}
		return children, nil
	case or:
		var children []fmt.Stringer
		for _, c := range node.children {
			children = append(children, c)
		}
		return children, nil
	case outR:
		return []fmt.Stringer{node.node}, nil
	case inR:
		return []fmt.Stringer{node.node}, nil
	}
	return nil, fmt.Errorf("unknown name, '%s'", node.name)
}
func (r *remover) boolIncoming(n fmt.Stringer) ([]fmt.Stringer, error) {
	node, ok := n.(removerNode)
	var incoming []fmt.Stringer
	for _, v := range r.allNodes {
		if !ok {
			if v.node != nil && v.node.String() == n.String() {
				incoming = append(incoming, v)
			}
			continue
		}
		for _, n := range v.children {
			if node.id == n.id {
				incoming = append(incoming, v)
			}
		}
	}
	return incoming, nil
}

var _ walder.NodeCreater = &remover{}

func (r *remover) NodeCreate(input fmt.Stringer) (fmt.Stringer, error) {
	if input == nil {
		return nil, fmt.Errorf("recieved nil value")
	}
	if r.allNodes == nil {
		r.allNodes = make(map[int]removerNode)
	}
	n, ok := boolNodes[input.String()]
	if !ok {
		return nil, fmt.Errorf("%s is unsupported - chose one of %#v", input, boolNodes) // TODO handle better
	}
	n.id = r.nextID()
	r.allNodes[n.id] = n
	return n, nil
}

var _ walder.EdgeCreater = &remover{}

func (r *remover) NodeUpdate(node fmt.Stringer) (fmt.Stringer, error) {
	if r.allNodes == nil {
		r.allNodes = make(map[int]removerNode)
	}
	if node == nil {
		return nil, fmt.Errorf("cant update nil node")
	}
	newNode, ok := node.(removerNode)
	if !ok {
		for _, n := range r.allNodes {
			if n.node != nil && n.node.String() == node.String() {
				return n, nil
			}
		}
		return nil, fmt.Errorf("this node '%#v' is unknown to this graph", node)
	}
	_, ok = r.allNodes[newNode.id]
	if !ok {
		return nil, fmt.Errorf("not initilised node %#v", newNode)
	}
	r.allNodes[newNode.id] = newNode
	return newNode, nil

}

func (r *remover) EdgeCreate(from, to fmt.Stringer) error {
	fromNode, ok := from.(removerNode)
	if !ok {
		return fmt.Errorf("cant create edge from %T", from)
	}
	if fromNode.name == outR {
		fromNode.node = to
		r.allNodes[fromNode.id] = fromNode
		r.removeTree = fromNode // TODO Fix
		return nil
	}
	toNode, ok := to.(removerNode)
	if !ok {
		return fmt.Errorf("cant create edge to %T", to)
	}
	switch fromNode.name {
	case not:
		fromNode.children = []removerNode{toNode}
	case and:
		fromNode.children = append(fromNode.children, toNode)
	case or:
		fromNode.children = append(fromNode.children, toNode)
	}
	r.allNodes[fromNode.id] = fromNode
	r.allNodes[toNode.id] = toNode
	r.removeTree = fromNode // TODO Fix
	return nil
}

func (r *remover) nextID() int {
	r.lastID++
	return r.lastID
}
func outReachable(g walder.GraphDirected, source, target fmt.Stringer) (bool, error) {
	var rec func(g walder.GraphDirected, source, target fmt.Stringer, seen map[string]struct{}) (bool, error)
	rec = func(g walder.GraphDirected, source, target fmt.Stringer, seen map[string]struct{}) (bool, error) {
		if g == nil {
			return false, fmt.Errorf("given graph is nil")
		}
		if source == nil {
			return false, fmt.Errorf("given source is nil")
		}
		if target == nil {
			return false, fmt.Errorf("given target is nil")
		}

		outR, err := g.Outgoing(source)
		if err != nil {
			return false, err
		}
		for _, o := range outR {
			if o == nil {
				continue
			}
			if o.String() == target.String() {
				return true, nil
			}
			if _, ok := seen[o.String()]; ok {
				return true, nil
			}
			seen[o.String()] = struct{}{}
			found, err := rec(g, o, target, seen)
			if err != nil {
				return found, err
			}
			if found {
				return true, nil
			}
		}
		return false, nil
	}
	seen := make(map[string]struct{})
	return rec(g, source, target, seen)
}
func inRechable(g walder.GraphDirected, source, target fmt.Stringer) (bool, error) {
	return outReachable(g, target, source)
}
