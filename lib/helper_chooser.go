package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type chooser struct {
	from     []fmt.Stringer
	execFunc func(w *Walder, node fmt.Stringer) error
	walder   *Walder
}

var _ internal = &chooser{}

func (c *chooser) renew(w *Walder) {
	c.walder = w
	return
}

var _ walder.Graph = &chooser{}

func (c *chooser) String() string {
	return "chooser"
}

func (c *chooser) HomeNodes() ([]fmt.Stringer, error) {
	return c.from, nil
}

var _ walder.GraphDirected = &chooser{}

func (c *chooser) Incoming(_ fmt.Stringer) ([]fmt.Stringer, error) {
	return []fmt.Stringer{}, nil
}
func (c *chooser) Outgoing(_ fmt.Stringer) ([]fmt.Stringer, error) {
	return []fmt.Stringer{}, nil
}

var _ walder.NodeOpener = &chooser{}

func (c *chooser) NodeOpen(nodes ...fmt.Stringer) (walder.Graph, error) {
	if len(c.from) != 0 {
		return c, fmt.Errorf("allready opened")
	}
	tmp := make([]fmt.Stringer, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		tmp = append(tmp, n)
	}
	if len(tmp) == 0 {
		return c, fmt.Errorf("nothing to open from")
	}
	c.from = tmp
	return c, nil
}

var _ walder.Executor = &chooser{}

func (c *chooser) Execute(node fmt.Stringer) error {
	if c.execFunc == nil {
		return fmt.Errorf("no execution function set")
	}
	if c.walder == nil {
		return fmt.Errorf("no pointer to walder set")
	}

	// TODO pop?

	err := c.execFunc(c.walder, node)

	return err
}
