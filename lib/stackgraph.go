package lib

import (
	"bytes"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	boxer "github.com/treilik/bubbleboxer"
	"github.com/treilik/walder"
)

const (
	inModusAddr   = "inModus"
	mainModusAddr = "mainModus"
	outModusAddr  = "outModus"
)

type graphHolder struct {
	focus string
	boxer *boxer.Boxer

	write      func(io.Reader) error
	graph      walder.Graph
	typer      *walder.Typer
	updateFunc func(*graphHolder, walder.Graph)

	lastInPosition  map[string]int
	lastOutPosition map[string]int
}

func (g *graphHolder) update() {
	t, ok := g.graph.(walder.Typer)
	if ok {
		g.typer = &t
	}
}

func (g *graphHolder) Write() error {
	gr, ok := g.graph.(walder.GetReader)
	if !ok {
		return fmt.Errorf("want %T, but got %T", gr, g.graph)
	}
	if g.write == nil {
		return fmt.Errorf("no write function set")
	}

	r, err := gr.GetReader()
	if err != nil {
		return fmt.Errorf("could not get reader to write graph because of: %s", err.Error())

	}
	err = g.write(r)
	if err != nil {
		return fmt.Errorf("while trying to write graph an error occured: %s", err.Error())
	}
	return nil
}

type stackTree struct {
	history [][]*graphHolder
	stack   []*graphHolder
}

var _ walder.GraphDirected = stackTree{}

type indexPath struct {
	path  []*graphHolder
	entry int
	level int
}

func (i indexPath) String() string {
	if i.path == nil {
		return "error"
	}
	b := bytes.Buffer{}
	for c, g := range i.path {
		if c > 0 {
			b.WriteString(", ")
		}
		b.WriteString((*g).graph.String())
	}
	return b.String()
}

func (s stackTree) String() string {
	return "stack tree"

}
func (s stackTree) HomeNodes() ([]fmt.Stringer, error) {
	if len(s.history) == 0 {
		return nil, fmt.Errorf("no history yet")
	}
	all := make([]fmt.Stringer, 0, len(s.history))
	for _, e := range s.history {

		all = append(all, indexPath{
			path:  e,
			level: len(e) - 1,
		})
	}

	return all, nil
}
func (s stackTree) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	ip, ok := node.(indexPath)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", ip, node)
	}
	if ip.level == 0 || len(ip.path) == 0 {
		return []fmt.Stringer{}, nil
	}
	parent := ip
	parent.level--
	return []fmt.Stringer{parent}, nil

}
func (s stackTree) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	ip, ok := node.(indexPath)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", ip, node)
	}

	if ip.level >= len(ip.path) {
		return nil, fmt.Errorf("level (%d) greater then path lenght (%d)", ip.level, len(ip.path))
	}
	parent := ip.path[ip.level]

	begin, end := -1, -1
	if ip.entry > 0 {
		for index := ip.entry - 1; index >= 0; index-- {
			path := s.history[index]
			if len(path) <= ip.level || path[ip.level] != parent {
				break
			}
			begin = index
		}
	}
	if ip.entry < len(s.history)-1 {
		for index := ip.entry + 1; index < len(s.history); index++ {
			path := s.history[index]
			if len(path) <= ip.level || path[ip.level] != parent {
				break
			}
			end = index
		}
	}
	if begin < 0 || end < 0 || end < begin {
		return []fmt.Stringer{}, nil
	}

	var children []fmt.Stringer
	childLevel := ip.level + 1
	for c := begin; c < end; c++ {
		children = append(children, indexPath{
			path:  s.history[c],
			entry: c,
			level: childLevel,
		})
	}
	return children, nil
}

func (s stackTree) peek() (*graphHolder, error) {
	if len(s.stack) == 0 {
		return nil, fmt.Errorf("cant peek into empty stack")
	}
	return (s.stack)[len(s.stack)-1], nil
}
func (s stackTree) pop() (*graphHolder, error) {
	s.history = append(s.history, s.stack)
	top, err := s.peek()
	if err != nil {
		return nil, err
	}
	(s.stack) = (s.stack)[:len(s.stack)-1]
	return top, nil
}
func (s stackTree) push(top *graphHolder) error {
	(s.stack) = append((s.stack), top)
	return nil
}

func (w *Walder) focus() string {
	return w.peek().focus
}

func (w *Walder) Pop() walder.Graph {
	return w.pop().graph
}
func (w *Walder) pop() *graphHolder {
	defer editCurrentGraph(w, func(i internal) error {
		i.renew(w)
		return nil
	})
	if len(w.stack.stack) == 1 {
		return (w.stack.stack)[0]
	}

	former := w.peek()

	if former.write != nil {
		w.addError(former.Write())
	}
	c, ok := former.graph.(walder.Closer)
	if ok {
		c.Close()
	}
	w.stack.stack = w.stack.stack[:len(w.stack.stack)-1]
	w.ui.ModelMap[graphAddr] = w.peek().boxer

	w.addError(w.peek().boxer.UpdateSize(tea.WindowSizeMsg{Width: w.width, Height: w.height}))
	m, ok := w.ui.ModelMap[graphAddr]
	if !ok {
		panic("no graph ui found")
	}
	switch graphBoxer := m.(type) {
	case boxer.Boxer:
		w.peek().boxer = &graphBoxer
	case *boxer.Boxer:
		w.peek().boxer = graphBoxer
	default:
		panic(fmt.Sprintf("unhandled type: %T", m))
	}

	return former
}
func (w *Walder) Push(newGraph walder.Graph, g ...graphHolder) error {
	if newGraph == nil {
		return fmt.Errorf("can't add nil value")
	}
	if len(g) > 1 {
		return fmt.Errorf("cant push multiple graphHolders at the same time") // TODO fix?
	}

	if i, ok := newGraph.(internal); ok {
		i.renew(w)
		newGraph = i
	}

	var b graphHolder
	if len(g) == 0 {
		b = newDirectedBoxer(newGraph)
		w.addError(b.boxer.UpdateSize(tea.WindowSizeMsg{Width: w.width, Height: w.height}))

		err := b.editList(mainAddr, func(l *holderList) error {
			h, err := newGraph.HomeNodes()
			if err != nil {
				return err
			}
			return l.ResetItems(h...)
		})
		if err != nil {
			return err
		}
		if _, ok := newGraph.(walder.GetReader); ok {
			storer := w.newStorer(newGraph.String())
			w.addError(err)
			if err == nil {
				b.write = func(r io.Reader) error {
					wc, err := storer()
					if err != nil {
						return err
					}
					_, err = io.Copy(wc, r)
					if err != nil {
						return err
					}
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
	} else {
		b = g[0]
	}

	if b.boxer == nil {
		return fmt.Errorf("recieved graphholder with <nil> boxer")
	}

	b.update()
	w.stack.stack = append(w.stack.stack, &b)
	w.ui.ModelMap[graphAddr] = b.boxer

	err := w.ui.UpdateSize(tea.WindowSizeMsg{Width: w.width, Height: w.height})
	w.addError(err)

	m, ok := w.ui.ModelMap[graphAddr]
	if !ok {
		panic("no graph ui found")
	}
	switch graphBoxer := m.(type) { // TODO fix
	case boxer.Boxer:
		w.peek().boxer = &graphBoxer
	case *boxer.Boxer:
		w.peek().boxer = graphBoxer
	default:
		panic(fmt.Sprintf("unhandled type: %T", m))
	}
	return nil
}
func (w *Walder) Replace(newGraph walder.Graph) (err error) {
	defer func() {
		if a := recover(); a != nil {
			err = fmt.Errorf("panic while replacing current graph: %v", a)
		}
	}()
	w.stack.history = append(w.stack.history, w.stack.stack)
	if newGraph == nil {
		return fmt.Errorf("can't add nil value")
	}
	if len(w.stack.stack) == 1 && !checkNav(newGraph) {
		return fmt.Errorf("can't replace base graph")
	}
	w.peek().graph = newGraph
	// TODO reset main list
	return nil
}

// checkNav checks recursivly if the given graph has a Navigator instance undelying.
func checkNav(wrapped walder.Graph) bool {
	if _, ok := wrapped.(*navigator); ok {
		return true
	}
	v, ok := wrapped.(walder.Meta)
	if !ok {
		return false
	}
	u, err := v.Get()
	if err != nil {
		return false
	}
	return checkNav(u)
}

func peek[T walder.Graph](w *Walder) (T, error) {
	top := w.peek()
	value, ok := top.graph.(T)
	if !ok {
		return value, fmt.Errorf("can't use '%T' as '%T'", top.graph, value)
	}
	return value, nil
}

func editCurrentGraph[T walder.Graph](w *Walder, editFunc func(T) error) error {
	v, err := peek[T](w)
	if err != nil {
		top, er := peek[walder.Meta](w)
		if er != nil {
			return err
		}
		defer func() {
			if a := recover(); a != nil {
				w.addError(fmt.Errorf("While editing the current graph, a panic occured: %#v", a))
			}
		}()
		changed, er := unwrap[T](top, editFunc)
		if er != nil {
			if _, ok := er.(nestingEnd); ok {
				return err
			}
			return er
		}
		return w.Replace(changed)
	}
	err = editFunc(v)
	if err != nil {
		return err
	}
	return w.Replace(v)
}

type nestingEnd error

func unwrap[T walder.Graph](g walder.Graph, editFunc func(T) error) (walder.Graph, error) {
	v, ok := g.(T)
	if ok {
		err := editFunc(v)
		return v, err
	}
	m, ok := g.(walder.Meta)
	if !ok {
		return g, nestingEnd(fmt.Errorf("end of nesting"))
	}
	u, err := m.Get()
	if err != nil {
		return g, err
	}
	changed, err := unwrap(u, editFunc)
	if err != nil {
		return g, err
	}
	err = m.Set(changed)
	return m, err

}

var (
	unfocused = termenv.Style{}
	focused   = termenv.Style{}.Reverse()
)

func (b *graphHolder) setFocus() {
	_ = b.editList(inAddr, func(l *holderList) error {
		l.CurrentStyle = unfocused
		if b.focus == inAddr {
			l.CurrentStyle = focused
		}
		return nil
	})
	_ = b.editList(mainAddr, func(l *holderList) error {
		l.CurrentStyle = unfocused
		if b.focus == mainAddr {
			l.CurrentStyle = focused
		}
		return nil
	})
	_ = b.editList(outAddr, func(l *holderList) error {
		l.CurrentStyle = unfocused
		if b.focus == outAddr {
			l.CurrentStyle = focused
		}
		return nil
	})
}
func stripErr(n boxer.Node, e error) boxer.Node { // TODO Handle error
	if e != nil {
		panic(e)
	}
	return n
}
