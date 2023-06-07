package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/treilik/walder"
)

type container struct {
	ID    int         `json:"id"`
	Name  string      `json:"name"`
	Nodes []container `json:"nodes"`
}

func (c container) String() string {
	return c.Name
}

type Sway struct {
	tree container
}

var _ walder.Dimensioner = Sway{}

func (s Sway) New() (walder.Graph, error) {
	return s, nil
}
func (s Sway) String() string {
	return "sway"
}

var _ walder.GraphDirectedTree = Sway{}

func (s Sway) HomeNodes() ([]fmt.Stringer, error) {
	cmd := exec.Command("swaymsg", "-t", "get_tree", "--raw")
	out, err := cmd.Output()
	var tree container
	json.Unmarshal(out, &tree)
	s.tree = tree
	return []fmt.Stringer{s.tree}, err
}
func (s Sway) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	parent, err := s.Parent(node)
	return []fmt.Stringer{parent}, err
}
func (s Sway) Parent(node fmt.Stringer) (fmt.Stringer, error) {
	current, ok := node.(container)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", current, node)
	}
	var parent fmt.Stringer
	visitor := func(path []fmt.Stringer) (abort bool, err error) {
		lenght := len(path)
		if lenght == 0 {
			return false, nil
		}
		last := path[lenght-1]
		n, ok := last.(container)
		if !ok {
			return true, fmt.Errorf("want %T, but got %T", n, last)
		}
		if n.ID != current.ID {
			return false, nil
		}
		if lenght == 1 {
			return true, fmt.Errorf("no parent of root node")
		}
		parent = path[lenght-2]
		return true, nil
	}
	cmd := exec.Command("swaymsg", "-t", "get_tree", "--raw")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var tree container
	json.Unmarshal(out, &tree)
	s.tree = tree

	return parent, DFS(s, visitor, s.tree)
}

func (s Sway) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	return s.Children(node)
}
func (s Sway) Children(node fmt.Stringer) ([]fmt.Stringer, error) {
	n, ok := node.(container)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", n, n)
	}

	if len(n.Nodes) == 1 && n.Nodes[0].Name == "" {
		// skip proxy child
		n = n.Nodes[0]
	}

	children := make([]fmt.Stringer, 0, len(n.Nodes))
	for _, child := range n.Nodes {
		children = append(children, child)
	}

	return children, nil
}

var _ walder.Executor = Sway{}

func (s Sway) Execute(node fmt.Stringer) error {
	c, ok := node.(container)
	if !ok {
		return fmt.Errorf("want %T, but got %T", c, node)
	}
	critiria := fmt.Sprintf("[con_id=\"%d\"]", c.ID)
	return exec.Command("swaymsg", critiria, "focus").Run()
}

var _ walder.NodeSwaper = Sway{}

func (s Sway) NodeSwap(first, second fmt.Stringer) error {
	a, ok := first.(container)
	if !ok {
		return fmt.Errorf("want %T, but got %T", a, first)
	}
	b, ok := second.(container)
	if !ok {
		return fmt.Errorf("want %T, but got %T", b, first)
	}

	cmd := exec.Command("swaymsg", "swap", "container", "with", "con_id", fmt.Sprintf("%d", a.ID), fmt.Sprintf("%d", b.ID))
	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer out.Close()
	err = cmd.Run()
	if err != nil {
		b, _ := io.ReadAll(out)
		return fmt.Errorf("%s: %s", err.Error(), b)
	}
	return nil
}
