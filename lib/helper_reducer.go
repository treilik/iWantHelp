package lib

import (
	"fmt"
	"regexp"

	"github.com/treilik/walder"
)

type directedAller interface {
	walder.GraphDirected
	walder.NodeAller
}

type reducer struct {
	origin  walder.GraphDirected
	sub     directedAller
	reducer func(walder.GraphDirected, fmt.Stringer) (fmt.Stringer, error)
}

var _ walder.Graph = &reducer{}

func (r *reducer) String() string {
	return "reducer"
}
func (r *reducer) HomeNodes() ([]fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("underling graph is nil")
	}
	homes, err := r.origin.HomeNodes()
	if err != nil {
		return nil, err
	}
	var filtered []fmt.Stringer
	for _, h := range homes {
		n, err := r.transform(h)
		if err != nil {
			return filtered, err
		}
		if n == nil {
			continue
		}
		filtered = append(filtered, n)
	}
	return filtered, nil
}

var _ walder.GraphDirected = &reducer{}

func (r *reducer) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("underling graph is nil")
	}
	homes, err := r.origin.Incoming(node)
	if err != nil {
		return nil, err
	}
	var filtered []fmt.Stringer
	for _, h := range homes {
		n, err := r.transform(h)
		if err != nil {
			return filtered, err
		}
		if n == nil {
			continue
		}
		filtered = append(filtered, n)
	}
	return filtered, nil

}
func (r *reducer) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("underling graph is nil")
	}
	homes, err := r.origin.Outgoing(node)
	if err != nil {
		return nil, err
	}
	var filtered []fmt.Stringer
	for _, h := range homes {
		n, err := r.transform(h)
		if err != nil {
			return filtered, err
		}
		if n == nil {
			continue
		}
		filtered = append(filtered, n)
	}
	return filtered, nil
}

func (r *reducer) transform(node fmt.Stringer) (fmt.Stringer, error) {
	if r.origin == nil {
		return nil, fmt.Errorf("underling graph is nil")
	}
	all, err := r.sub.NodeAll()
	if err != nil {
		return nil, err
	}
	for _, a := range all {
		ok, err := regexp.MatchString(a.String(), node.String())
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		inOK, err := r.hasIncoming(a, node)
		if err != nil {
			return nil, err
		}
		outOK, err := r.hasOutgoing(a, node)
		if err != nil {
			return nil, err
		}
		if inOK && outOK {
			return r.reducer(r.origin, node)
		}

	}

	return node, nil
}

func (r *reducer) hasOutgoing(subNode, node fmt.Stringer) (bool, error) {
	subOut, err := r.sub.Outgoing(subNode)
	if err != nil {
		return false, err
	}
	if len(subOut) == 0 {
		return true, nil
	}
	out, err := r.origin.Outgoing(node)
	if err != nil {
		return false, err
	}

outer:
	for _, s := range subOut {
		for _, o := range out {
			ok, err := regexp.MatchString(s.String(), o.String())
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}

			// deep first search
			ok, err = r.hasOutgoing(s, o)
			if err != nil {
				return false, nil
			}
			if !ok {
				continue
			}
			// found a matching subsubgraph
			continue outer
		}
		return false, nil
	}
	return true, nil
}
func (r *reducer) hasIncoming(subNode, node fmt.Stringer) (bool, error) {
	subOut, err := r.sub.Incoming(subNode)
	if err != nil {
		return false, err
	}
	if len(subOut) == 0 {
		return true, nil
	}
	out, err := r.origin.Incoming(node)
	if err != nil {
		return false, err
	}

outer:
	for _, s := range subOut {
		for _, o := range out {
			ok, err := regexp.MatchString(s.String(), o.String())
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}

			// deep first search
			ok, err = r.hasIncoming(s, o)
			if err != nil {
				return false, nil
			}
			if !ok {
				continue
			}
			// found a matching subsubgraph
			continue outer
		}
		return false, nil
	}
	return true, nil
}
