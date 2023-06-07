package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type cmdStack struct {
	stack []*command
}

func (s *cmdStack) peek() (*command, error) {
	if len(s.stack) == 0 {
		return nil, fmt.Errorf("empty stack")
	}
	top := s.stack[len(s.stack)-1]
	return top, nil
}

func (s *cmdStack) push(c *command) error {
	if c == nil {
		return fmt.Errorf("recieved nil value")
	}
	s.stack = append(s.stack, c)
	return nil
}

type commandGraph struct {
	commands map[string]command
	walder   *Walder
}

var _ walder.Graph = &commandGraph{}

func (g *commandGraph) String() string {
	return "commands"
}

func (g *commandGraph) HomeNodes() ([]fmt.Stringer, error) {
	if g.commands == nil {
		return nil, fmt.Errorf("no commands set")
	}
	stringerList := make([]fmt.Stringer, 0, len(g.commands))
	for _, v := range g.commands {
		stringerList = append(stringerList, v) // TODO cache
	}

	return sortStringer(stringerList), nil
}

var _ internal = &commandGraph{}

func (g *commandGraph) renew(w *Walder) {
	// TODO is here a race conflict with w.bindings getting set ?
	if g.commands == nil {
		g.commands = make(map[string]command, w.bindings.bindings.Order())
		nodes, err := w.bindings.nodes()
		if err != nil {
			panic(err) // TODO
		}
		for _, b := range nodes {
			if b.Cmd.run == nil {
				continue
			}
			g.commands[b.Cmd.Name] = b.Cmd
		}
	}
	g.walder = w
	return
}

var _ walder.GraphDirected = &commandGraph{}

func (g commandGraph) Incoming(_ fmt.Stringer) ([]fmt.Stringer, error) {
	return []fmt.Stringer{}, nil
}
func (g commandGraph) Outgoing(_ fmt.Stringer) ([]fmt.Stringer, error) {
	return []fmt.Stringer{}, nil
}

var _ walder.Executor = &commandGraph{}

func (g *commandGraph) Execute(node fmt.Stringer) error {
	if g.walder == nil {
		return fmt.Errorf("walder not set")
	}
	cmd, ok := node.(command)
	if !ok {
		return fmt.Errorf("want '%#v', but got '%#v'", cmd, node)
	}
	g.walder.pop() // TODO change
	// TODO set cursor and so on?
	cmd.walder = g.walder
	return cmd.run(&cmd) // TODO protect walder from change wenn error occurs
}

func newCommand() *command {
	cmd := &command{}
	cmd.init()
	return cmd
}

type command struct {
	Name         string
	Description  string
	run          func(*command) error
	repeatAmount string

	pauseReason string

	pausing  chan struct{}
	resuming chan struct{}
	done     chan struct{}

	timeOuted bool

	walder *Walder

	err error
}

func (c command) String() string {
	if c.pauseReason != "" {
		return fmt.Sprintf("%s\npausing to:\n%s", c.Name, c.pauseReason)
	}
	return c.Name
}

func (c *command) init() {
	c.pausing = make(chan struct{})
	c.resuming = make(chan struct{})
	c.done = make(chan struct{})
}

func (c *command) Run(w *Walder) {
	if c.pausing == nil {
		return
	}
	if c.resuming == nil {
		return
	}
	if c.done == nil {
		return
	}

	if w == nil {
		c.err = fmt.Errorf("recived nil walder")
		return
	}
	if c.run == nil {
		c.err = fmt.Errorf("no run function set")
		return
	}

	c.walder = w

	defer close(c.pausing)
	defer close(c.resuming)
	defer close(c.done)

	defer func() {
		if c.walder.peek().write != nil {
			c.walder.addError(c.walder.peek().Write())
		}
	}()

	c.walder.cmdStack.push(c)

	c.walder.addError(c.run(c))

	for i, cmd := range c.walder.cmdStack.stack {
		if cmd.done != c.done {
			continue
		}
		if cmd.timeOuted {
			continue
		}
		var rest []*command
		if len(c.walder.cmdStack.stack) > i {
			rest = c.walder.cmdStack.stack[i+1:]
		}
		c.walder.cmdStack.stack = append(c.walder.cmdStack.stack[:i], rest...)
	}
}

func (c *command) pause(reason string) {
	c.pauseReason = reason
	defer func() {
		c.pauseReason = ""
	}()

	c.walder = nil
	c.pausing <- struct{}{}
	<-c.resuming
	if c.walder == nil {
		panic("walder was not set again")
	}
	if c.run == nil {
		c.walder.addError(fmt.Errorf("no run function set"))
		return
	}
}

func (c *command) resume(w *Walder) {
	c.walder = w
	select {
	case timeOuted := <-c.pausing:
		_ = timeOuted
	default:
	}
	c.resuming <- struct{}{}
}

func (c *command) isClosed() bool {
	var closed bool
	select {
	case _, closed = <-c.done:
	default:
	}
	return closed
}
