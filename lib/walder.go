package lib

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	boxer "github.com/treilik/bubbleboxer"
	lister "github.com/treilik/bubblelister"
	"github.com/treilik/walder"
)

const (
	inAddr   = "in"
	mainAddr = "main"
	outAddr  = "out"

	stackAddr = "stack"
	graphAddr = "graph"
	cmdAddr   = "cmd"

	errorAddr = "error"
	inputAddr = "input"
)

type syncer struct {
	cmd *command
}

type internal interface {
	walder.Graph
	renew(w *Walder) // TODO add error
}

// NewWalder returns Walder fully setup up
func NewWalder() *Walder {
	w := &Walder{}
	w.cmdTimeout = time.Second

	w.Syncer = make(chan syncer)

	w.cmdStack = &cmdStack{}

	c := &cmdDag{}
	err := c.Add(commandList)
	if err != nil {
		panic(err)
	}
	w.bindings = c

	n := &navigator{walder: w}
	g := newTreeModus(n)
	_ = g.editList(mainAddr, func(l *holderList) error {
		nodes, err := n.HomeNodes()
		if err != nil {
			panic(fmt.Sprintf("error while startup: %s\n", err))
		}
		nilerr := l.AddItems(nodes...)
		if nilerr != nil {
			panic("nil values where non should be")
		}
		return nil
	})

	w.ui = boxer.Boxer{}
	w.ui.ModelMap = make(map[string]tea.Model)
	w.ui.LayoutTree = boxer.Node{
		// root node
		VerticalStacked: true,
		SizeFunc:        func(_ boxer.Node, height int) []int { return []int{height - 1, 1} },
		Children: []boxer.Node{
			{
				// list node
				SizeFunc: func(_ boxer.Node, width int) []int {
					stackWidth := 20
					tasksWidth := 20
					graphWidth := width - (stackWidth + tasksWidth)
					if graphWidth < 0 {
						stackWidth = width / 3
						tasksWidth = width / 3
						graphWidth = width - (stackWidth + tasksWidth)
					}
					return []int{stackWidth, graphWidth, tasksWidth}
				},
				Children: []boxer.Node{
					stripErr(w.ui.CreateLeaf(stackAddr, lister.NewModel())),
					stripErr(w.ui.CreateLeaf(graphAddr, g.boxer)),
					stripErr(w.ui.CreateLeaf(cmdAddr, lister.NewModel())),
				},
			},
			{
				// input and error node
				Children: []boxer.Node{
					stripErr(w.ui.CreateLeaf(inputAddr, input{textinput.NewModel()})),
					stripErr(w.ui.CreateLeaf(errorAddr, lister.NewModel())),
				},
			},
		},
	}

	w.addError(w.Push(n, g))

	return w
}

// Walder is a graph-editor
type Walder struct {
	inputFunc func(w *Walder, input string) error

	ui    boxer.Boxer
	stack stackTree

	bindings  *cmdDag
	keyBuffer []tea.KeyMsg

	cmdStack *cmdStack

	dims   []walder.Dimensioner
	graphs []walder.Graph

	width  int
	height int

	Syncer chan syncer

	cmdTimeout time.Duration
}

func (w *Walder) newStorer(name string) func() (io.WriteCloser, error) {
	now := time.Now().UnixMicro()
	return func() (io.WriteCloser, error) {
		home := os.Getenv("HOME")
		return os.Create(fmt.Sprintf("%s/.walder/%d_%s", home, now, name))
	}
}

func (w *Walder) waitForCommand(cmd *command) error {
	if cmd == nil {
		return fmt.Errorf("cant wait for <nil> command")
	}

	if cmd.run == nil {
		return fmt.Errorf("command was not correctly set up")
	}
	if cmd.pausing == nil {
		return fmt.Errorf("command was not correctly set up")
	}
	if cmd.resuming == nil {
		return fmt.Errorf("command was not correctly set up")
	}
	if cmd.done == nil {
		return fmt.Errorf("command was not correctly set up")
	}

	timeout := time.After(w.cmdTimeout)
	select {
	case <-cmd.pausing:
	case <-cmd.done:
	case <-timeout:
		cmd.timeOuted = true
	}
	return cmd.err
}

var _ tea.Model = &Walder{}

// Init satisfies the bubbletea.Model interface and does nothing else
func (w *Walder) Init() tea.Cmd { return nil }

// View renders the graph editor for the user
func (w *Walder) View() (returnString string) {
	defer func() {
		if r := recover(); r != nil {
			// named return values can be changed within defered functions
			returnString = fmt.Sprintf("error while rendering: %#v\n", r)
		}
	}()
	var stackEntries []fmt.Stringer

	for _, g := range w.stack.stack {
		stackEntries = append(stackEntries, stringer(g.graph.String()))
	}

	w.ui.EditLeaf(cmdAddr, func(m tea.Model) (tea.Model, error) {
		if m == nil {
			return m, fmt.Errorf("got <nil>")
		}
		l, ok := m.(lister.Model)
		if !ok {
			return nil, fmt.Errorf("want %T, but got %T", l, m)
		}
		if w.cmdStack == nil {
			w.cmdStack = &cmdStack{}
		}
		var stringers []fmt.Stringer
		for _, s := range w.cmdStack.stack {
			stringers = append(stringers, s)
		}
		l.ResetItems(stringers...)
		return l, nil
	})
	w.ui.EditLeaf(stackAddr, func(m tea.Model) (tea.Model, error) {
		l, ok := m.(lister.Model)
		if !ok {
			return nil, fmt.Errorf("want %T, but got %T", l, m)
		}
		l.ResetItems(stackEntries...)
		return l, nil
	})

	g := w.peek()
	g.update()

	ui := w.ui
	ui.ModelMap[graphAddr] = g.boxer
	ui.UpdateSize(tea.WindowSizeMsg{Width: w.width, Height: w.height})

	returnString = ui.View()
	return returnString
}

// Update changes the graph-editors internal state according to the events or Messages received
func (w *Walder) Update(msg tea.Msg) (m tea.Model, _ tea.Cmd) {
	var focus string
	defer func() { m = w }()
	defer func() { w.peek().focus = focus }()
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		w.addError(fmt.Errorf("after message '%#v' caught panic: '%#v'", msg, r))
	}()
	// keep user interface up to date with changes
	defer w.updateUI()
	// keep keybindings up to date
	defer w.updateBindings() // TODO remove since keybindings can be change in walder graph
	defer func() { focus = w.peek().focus }()

	g := w.peek()
	w.ui.ModelMap[graphAddr] = g.boxer

	// TODO reset focus if graph changes

	// give internal graphs access to everything
	_ = editCurrentGraph(w, func(i internal) error {
		i.renew(w)
		return nil
	})
	w.peek().update()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" {
			return w, tea.Quit
		}
		if w.inputFunc != nil {
			if key == "esc" {
				_, err := w.deactivateInput() // TODO set last func here
				w.addError(err)
				return w, nil
			}
			if key == "enter" {
				inputFunc, err := w.deactivateInput()
				if err != nil {
					w.addError(err)
					return w, nil
				}

				newWalder := w // TODO is here a deep copy of Walder necessary
				err = inputFunc(newWalder)
				if err != nil {
					w.addError(err)
					return w, nil
				}

				// replace walder through the changed instance
				return newWalder, nil
			}
			model, ok := w.ui.ModelMap[inputAddr]
			if !ok {
				w.addError(fmt.Errorf("no input Model in LayoutTree"))
				return w, nil
			}

			inputModel, ok := model.(input)
			if !ok {
				w.addError(fmt.Errorf("wrong input Model in LayoutTree"))
				return w, nil
			}
			model, _ = inputModel.Update(msg) // TODO handle cmd differnt?
			w.ui.ModelMap[inputAddr] = model
			return w, nil
		}

		tmpBuffer := make([]tea.KeyMsg, len(w.keyBuffer))
		copy(tmpBuffer, w.keyBuffer)
		tmpBuffer = append(tmpBuffer, msg)

		av, err := w.bindings.Filter(tmpBuffer)
		if err != nil {
			w.addError(err)
			return w, nil
		}

		if len(av) > 0 {
			if len(av) == 1 {
				cmd := &av[0].Cmd

				cmd.init()
				go cmd.Run(w)

				err = w.waitForCommand(cmd)
				w.addError(err)

				return w, nil
			}
			w.keyBuffer = tmpBuffer
			return w, nil
		}

		// reset key buffer since no command was found
		w.keyBuffer = []tea.KeyMsg{}

		switch key {
		case "?":
			cmd, err := w.cmdStack.peek()
			if err != nil {
				w.addError(err)
				return w, nil
			}
			cmd.timeOuted = false
			select {
			case <-cmd.done:
				if w == cmd.walder {
					w.addError(fmt.Errorf("command done but has same state")) // TODO
				}
				return cmd.walder, nil
			default:
			}

			select {
			case <-cmd.pausing:
				cmd.resume(w)
				return w, nil
			default:
			}
			w.addError(fmt.Errorf("not yet done"))
			return w, nil
		case "!":
			newCmd := newCommand()
			w.addError(w.cmdStack.push(newCmd))
			return w, nil
		case "ctrl+c", "q":
			return w, tea.Quit
		case "esc":
			w.peek().focus = mainAddr
			w.addError(w.ui.EditLeaf(errorAddr, func(m tea.Model) (tea.Model, error) {
				l, ok := m.(lister.Model)
				if !ok {
					return nil, fmt.Errorf("want %T, but got %T", l, m)
				}
				l.ResetItems()
				return l, nil
			}))
		case "enter":
			cur, err := w.peek().getCursorItem()
			if err != nil {
				w.addError(err)
				return w, nil
			}

			c, ok := cur.(command)
			if !ok {
				e, err := peek[walder.Executor](w)
				if err != nil {
					w.addError(err)
					return w, nil
				}
				w.addError(e.Execute(cur))
				return w, nil
			}
			w.pop()

			cmd := &c

			cmd.init()
			go cmd.Run(w)

			err = w.waitForCommand(cmd)
			w.addError(err)

			return w, nil

		case "A":
			last, err := w.cmdStack.peek()
			if err != nil {
				w.addError(err)
				return w, nil
			}

			if last.isClosed() {
				return w, nil
			}

			last.resume(w)

			err = w.waitForCommand(last)
			w.addError()

			return w, nil
		case "b":
			w.pop()
			return w, nil
		case "c":
			err := w.Push(&commandGraph{})
			w.addError(err)
			return w, nil
		case "up":
			w.addError(w.peek().editList(w.peek().focus, func(l *holderList) error {
				_, _ = l.MoveCursor(-1)
				return nil
			}))
			return w, nil
		case "down":
			w.addError(w.peek().editList(w.peek().focus, func(l *holderList) error {
				_, _ = l.MoveCursor(1)
				return nil
			}))
			return w, nil
		case "left":
			w.addError(w.peek().moveIncoming())
			return w, nil
		case "right":
			w.addError(w.peek().moveOutgoing())
			return w, nil
		case "alt+[H":
			w.addError(w.peek().editList(mainAddr, func(l *holderList) error {
				return l.Top()
			}))
			return w, nil
		case "alt+[F":
			w.addError(w.peek().editList(mainAddr, func(l *holderList) error {
				return l.Bottom()
			}))
			return w, nil
		case "D":
			cur, err := w.peek().getCursorItem()
			if err != nil {
				w.addError(err)
				return w, nil
			}

			g, err := peek[walder.Graph](w)
			if err != nil {
				w.addError(err)
				return w, nil
			}
			dim, ok := g.(walder.Dimensions)
			if !ok {
				w.addError(fmt.Errorf("want %s, but got %T", dimensionsString, g))
				return w, nil
			}
			w.addError(w.Push(changer{
				dim:  dim,
				root: origin(cur),
			}))
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			cmd, _ := w.cmdStack.peek()
			if cmd != nil && cmd.run != nil {
				w.cmdStack.push(&command{repeatAmount: key})
			}
			cmd.repeatAmount += key
		case "alt+left":
			w.peek().focus = inAddr
		case "alt+right":
			w.peek().focus = outAddr
		default:
			w.addError(fmt.Errorf("unbound key: '%s'", key))
		}
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		err := w.ui.UpdateSize(msg)
		b := w.ui.ModelMap[graphAddr].(boxer.Boxer)

		// write changed graph boxer back into ui for render // TODO find better way
		w.peek().boxer = &b
		w.addError(err)
		return w, nil
	default:
		w.addError(fmt.Errorf("unknown message: '%#v'", msg))
	}
	return w, nil
}

func (w *Walder) DimensionRegister(dims ...walder.Dimensioner) {
	newDims := make([]walder.Dimensioner, 0, len(dims))
	for _, d := range dims {
		if d == nil {
			w.addError(fmt.Errorf("a dimension was not added because it was nil"))
			continue
		}
		newDims = append(newDims, d)
	}
	w.dims = append(w.dims, newDims...)
}
func (w *Walder) GraphRegister(graphs ...walder.Graph) {
	for _, a := range graphs {
		if a == nil {
			w.addError(fmt.Errorf("a graph adapter was not added because it was nil"))
			continue
		}
		w.graphs = append(graphs, a)
	}
}

// TODO add example for keybinding reader content

// ReadBindings read from the given reader the keybindings which shoul be used.
// this overwrites former used keybindings.
func (w *Walder) ReadBindings(reader io.Reader) error {
	if reader == nil {
		return fmt.Errorf("cant read from nil")
	}

	cd := &cmdDag{}
	c, err := cd.Open(reader)
	if err != nil {
		return fmt.Errorf("while reading the keybindings a error occured: %w", err)
	}

	w.bindings = c.(*cmdDag)

	return nil
}
func (w *Walder) addError(errList ...error) {
	stringerList := make([]fmt.Stringer, 0, len(errList))
	for _, err := range errList {
		if err == nil {
			continue
		}
		stringerList = append(stringerList, errStringer{err})
	}
	if len(stringerList) == 0 {
		return
	}

	errModel := w.ui.ModelMap[errorAddr]
	errLister := errModel.(lister.Model)

	_ = errLister.ResetItems(stringerList...)

	w.ui.ModelMap[errorAddr] = errLister
}

type errStringer struct {
	err error
}

func (e errStringer) String() string {
	return e.err.Error()
}

func (w *Walder) updateBindings() {
	bindings, err := peek[*cmdDag](w)
	if err != nil {
		return
	}
	w.bindings = bindings
}

func (w *Walder) updateUI() {
	top := w.peek()
	if top.focus != mainAddr {
		return
	}
	top.updateFunc(top, top.graph)
}

func (w *Walder) peek() *graphHolder {
	if len(w.stack.stack) == 0 {
		nav := newDirectedBoxer(&navigator{})
		w.stack.stack = []*graphHolder{&nav}
	}
	return (w.stack.stack)[len(w.stack.stack)-1]
}
