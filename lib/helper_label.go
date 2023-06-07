package lib

import (
	"fmt"
	"strings"

	"github.com/treilik/walder"
)

var _ walder.GraphDirected = labelWrapper{}

type directedLabler interface {
	walder.NodeLabeler
	walder.GraphDirected
}
type labelWrapper struct {
	origin directedLabler
}

func (l labelWrapper) String() string {
	if l.origin == nil {
		return "nothing to wrap"
	}
	return fmt.Sprintf("wrapping: %s", l.origin.String())
}
func (l labelWrapper) HomeNodes() ([]fmt.Stringer, error) {
	if l.origin == nil {
		return nil, fmt.Errorf("nothing to wrap")
	}
	nodes, err := l.origin.HomeNodes()
	if err != nil {
		return nil, err
	}
	return l.toHolder(nodes)
}
func (l labelWrapper) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	if l.origin == nil {
		return nil, fmt.Errorf("nothing to wrap")
	}

	n, ok := node.(labelHolder)
	if !ok {
		return nil, fmt.Errorf("not of this graph")
	}
	nodes, err := l.origin.Incoming(n.node)
	if err != nil {
		return nil, err
	}
	return l.toHolder(nodes)
}
func (l labelWrapper) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	if l.origin == nil {
		return nil, fmt.Errorf("nothing to wrap")
	}

	n, ok := node.(labelHolder)
	if !ok {
		return nil, fmt.Errorf("not of this graph")
	}
	nodes, err := l.origin.Outgoing(n.node)
	if err != nil {
		return nil, err
	}
	return l.toHolder(nodes)
}

type labelHolder struct {
	node   fmt.Stringer
	labels [][2]string
}

func (l labelHolder) String() string {
	if l.node == nil || l.labels == nil {
		return ""
	}
	all := make([]string, 0, len(l.labels)+1)
	all = append(all, l.node.String())
	for _, v := range l.labels {
		all = append(all, fmt.Sprintf("  %s: %v", v[0], v[1]))
	}
	return strings.Join(all, "\n") // TODO change to constant
}

func (l labelWrapper) toHolder(nodes []fmt.Stringer) ([]fmt.Stringer, error) {
	holders := make([]fmt.Stringer, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			return holders, fmt.Errorf("a underlying node was nil")
		}
		labels, err := l.origin.NodeLabels(n)
		if err != nil {
			return holders, err
		}
		holders = append(holders, labelHolder{node: n, labels: labels})
	}
	return holders, nil
}
