package lib

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/treilik/walder"
)

type String struct{}

func (s String) String() string {
	return "string"
}

func (s String) New() (walder.Graph, error) {
	return stringerGraph{}, nil
}

var _ walder.OpenReader = String{}

func (s String) Open(r io.Reader) (walder.Graph, error) {
	all, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	New := make(map[int]string)
	for i, line := range strings.Split(string(all), "\n") {
		New[i+1] = line
	}
	return stringGrapher(New), nil
}

type stringNode struct {
	index int
	str   string
}

func (sn stringNode) String() string {
	return sn.str
}

type stringGrapher map[int]string

var _ walder.Graph = stringGrapher{}

func (l stringGrapher) String() string {
	return "string"
}
func (l stringGrapher) HomeNodes() ([]fmt.Stringer, error) {
	return l.NodeAll()
}

var _ walder.NodeAller = stringGrapher{}

func (l stringGrapher) NodeAll() ([]fmt.Stringer, error) {
	lines := make([]fmt.Stringer, len(l))
	for k, v := range l {
		lines[k-1] = stringNode{str: v, index: k}
	}
	return lines, nil
}

var _ walder.NodeSwaper = stringGrapher{}

func (l stringGrapher) NodeSwap(a, b fmt.Stringer) error {
	first, ok := a.(stringNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", first, a)
	}
	second, ok := b.(stringNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", second, b)
	}
	if n, ok := l[first.index]; !ok || n != first.str {
		return fmt.Errorf("node %#v not found", first)
	}
	if n, ok := l[second.index]; !ok || n != second.str {
		return fmt.Errorf("node %#v not found", first)
	}
	l[first.index], l[second.index] = second.str, first.str
	return nil
}

var _ walder.GetReader = stringGrapher{}

func (l stringGrapher) GetReader() (io.Reader, error) {
	all, err := l.NodeAll()
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	for _, a := range all {
		b.WriteString(a.String())
	}
	return &b, nil
}
