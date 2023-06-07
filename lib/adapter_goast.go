package lib

import (
	"bytes"
	"fmt"
	"io"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/treilik/walder"
)

type GoastDim struct{}

var _ walder.Dimensioner = GoastDim{}

func (d GoastDim) String() string {
	return "go ast"
}
func (d GoastDim) New() (walder.Graph, error) {
	return nil, fmt.Errorf("not yet implemented - use Open") // TODO
}

var _ walder.OpenReader = GoastDim{}

func (d GoastDim) Open(r io.Reader) (walder.Graph, error) {
	all, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ag := &astGraph{}
	ag.filecontent = all

	fs := token.NewFileSet()
	ag.fileSet = fs
	a, err := parser.ParseFile(fs, "", ag.filecontent,
		parser.AllErrors|parser.ParseComments,
	)
	ag.f = a
	ag.tree = make(map[astNode][]astNode)
	return ag, err
}

type astNode struct {
	n    ast.Node
	text string
	id   int
}

func (n astNode) String() string {
	if n.n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T: %s", n.n, n.text)
}

type astGraph struct {
	fileSet     *token.FileSet
	filecontent []byte
	f           *ast.File
	tree        map[astNode][]astNode
	path        []astNode
	last        int
}

func (a astGraph) String() string {
	return "ast"
}
func (a *astGraph) HomeNodes() ([]fmt.Stringer, error) {
	ast.Walk(a, a.f)
	var stringers []fmt.Stringer
	for n, _ := range a.tree {
		stringers = append(stringers, n)
	}
	return stringers, nil
}

var _ walder.GraphDirected = &astGraph{}

func (a *astGraph) Incoming(n fmt.Stringer) ([]fmt.Stringer, error) {
	astNode, ok := n.(astNode)
	if !ok {
		return nil, fmt.Errorf("got %T want %T", n, astNode)
	}
	var strs []fmt.Stringer
out:
	for k, v := range a.tree {
		for _, e := range v {
			if e.id == astNode.id {
				strs = append(strs, k)
				// only one parent in a tree
				break out
			}
		}
	}
	return strs, nil
}
func (a *astGraph) Outgoing(n fmt.Stringer) ([]fmt.Stringer, error) {
	astNode, ok := n.(astNode)
	if !ok {
		return nil, fmt.Errorf("got %T want %T", n, astNode)
	}
	out, ok := a.tree[astNode]
	if !ok {
		return nil, fmt.Errorf("'%#v' isnot in this graph", astNode)
	}
	var strs []fmt.Stringer
	for _, o := range out {
		strs = append(strs, o)
	}
	return strs, nil
}

var _ ast.Visitor = &astGraph{}

func (ag *astGraph) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		// deepfirstsearch leafs this subtree thus assend level
		if len(ag.path) > 0 {
			ag.path = ag.path[:len(ag.path)-1]
		}
		return ag
	}
	text := string(ag.filecontent[n.Pos()-1 : n.End()])

	newNode := astNode{
		n:    n,
		text: text,
		id:   ag.nextID(),
	}

	ag.tree[newNode] = []astNode{}

	if len(ag.path) > 0 {
		parent := ag.path[len(ag.path)-1]
		ag.tree[parent] = append(ag.tree[parent], newNode)
	}
	ag.path = append(ag.path, newNode)
	return ag
}

var _ walder.GetReader = &astGraph{}

func (ag *astGraph) GetReader() (io.Reader, error) {
	b := &bytes.Buffer{}
	printer.Fprint(b, ag.fileSet, ag.f)
	return b, nil
}

var _ walder.EdgeMover = &astGraph{}

func (ag *astGraph) EdgeMove(toMove, from, to fmt.Stringer) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %#v", r)
		}
	}()
	m, ok := toMove.(astNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", m, toMove)
	}
	f, ok := from.(astNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", f, from)
	}
	t, ok := to.(astNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", t, to)
	}
	var foundM, foundF, foundT bool
	var moveScope bool
	astutil.Apply(
		ag.f,
		func(c *astutil.Cursor) bool {
			cur := c.Node()
			if cur == nil {
				return true
			}
			if equal(cur, m.n) {
				moveScope = true
			}
			return true
		},
		func(c *astutil.Cursor) bool {
			defer func() { moveScope = false }()
			cur := c.Node()
			if cur == nil {
				return true
			}
			if equal(cur, m.n) {
				foundM = true
			}
			if equal(cur, f.n) {
				foundF = true
			}
			if equal(cur, t.n) {
				if moveScope {
					err = fmt.Errorf("cant move subtree into it self")
				}
				foundT = true
			}

			return true
		},
	)
	if err != nil {
		return err
	}
	if !(foundM && foundF && foundT) {
		return fmt.Errorf("did not found one of the nodes")
	}
	astutil.Apply(ag.f, nil, func(c *astutil.Cursor) bool {
		cur := c.Node()
		if cur == nil {
			return true
		}
		if equal(cur, m.n) {
			c.Delete()
		}
		if equal(c.Parent(), t.n) && c.Index() == 0 {
			c.InsertBefore(m.n)
		}
		return true
	})
	return nil
}

func equal(a, b ast.Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	if b == nil {
		return false
	}

	return a.Pos() == b.Pos() && a.End() == b.End()
}

var _ walder.NodeSwaper = &astGraph{}

func (ag *astGraph) NodeSwap(first, second fmt.Stringer) error {
	f, ok := first.(astNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", f, first)
	}
	s, ok := second.(astNode)
	if !ok {
		return fmt.Errorf("want %T, but got %T", s, second)
	}
	astutil.Apply(ag.f, nil, func(c *astutil.Cursor) bool {
		cur := c.Node()
		if cur == nil {
			return true
		}
		if cur.Pos() == f.n.Pos() && cur.End() == f.n.End() {
			c.Replace(s.n)
			return true
		}
		if cur.Pos() == s.n.Pos() && cur.End() == s.n.End() {
			c.Replace(f.n)
			return true
		}
		return true
	})
	return nil
}
func (ag *astGraph) nextID() int {
	ag.last++
	return ag.last
}
