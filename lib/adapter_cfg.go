package lib

import (
	"fmt"
	"go/ast"

	"github.com/treilik/walder"
	"golang.org/x/tools/go/cfg"
)

type CfgDim struct{}

var _ walder.NodeOpener = CfgDim{}
var _ walder.Dimensioner = CfgDim{}

func (d CfgDim) String() string {
	return "Controll Flow Graph of golang"
}

func (d CfgDim) New() (walder.Graph, error) {
	return nil, fmt.Errorf("not yet implemented")
}
func (d CfgDim) NodeOpen(nodes ...fmt.Stringer) (walder.Graph, error) {
	if len(nodes) != 1 {
		return nil, fmt.Errorf("want exactly one node")
	}
	n := nodes[0]
	a, ok := n.(astNode)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", a, n)
	}
	fd, ok := a.n.(*ast.FuncDecl)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", fd, a.n)
	}
	c := cfg.New(fd.Body, func(*ast.CallExpr) bool { return false })
	cfg := gocfg{fnc: fd,
		cfg: c}

	return cfg, nil
}

var _ walder.GraphDirected = gocfg{}

type gocfg struct {
	fnc *ast.FuncDecl
	cfg *cfg.CFG
}

func (g gocfg) String() string {
	return g.fnc.Name.String()
}
func (g gocfg) HomeNodes() ([]fmt.Stringer, error) {
	var nodes []fmt.Stringer
	for _, b := range g.cfg.Blocks {
		nodes = append(nodes, b)
	}
	return nodes, nil
}
func (g gocfg) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	n, ok := node.(*cfg.Block)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", n, node)
	}
	var pred []fmt.Stringer
	for _, b := range g.cfg.Blocks {
		for _, s := range b.Succs {
			if s.Index == n.Index {
				pred = append(pred, b)
			}
		}
	}
	return pred, nil
}
func (g gocfg) Outgoing(n fmt.Stringer) ([]fmt.Stringer, error) {
	b, ok := n.(*cfg.Block)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", b, n)
	}
	var succs []fmt.Stringer
	for _, s := range b.Succs {
		succs = append(succs, s)
	}

	return succs, nil
}
