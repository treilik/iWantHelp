package lib

import (
	"fmt"
	"strconv"

	"github.com/treilik/walder"
)

const (
	graphString            = "Graph"
	graphNeighborsString   = "GraphNeighbors"
	graphIncomingString    = "GraphIncoming"
	graphOutgoingString    = "GraphOutgoing"
	graphDirectedString    = "GraphDirected"
	graphEqualerString     = "GraphEqualer"
	graphLesserString      = "GraphLesser"
	graphConstrainerString = "GraphConstrainer"
	getReaderString        = "GetReader"
	nodeReaderString       = "NodeReader"
	nodeCreaterString      = "NodeCreater"
	nodeFromCreaterString  = "NodeFromCreater"
	nodeToCreaterString    = "NodeToCreater"
	nodeUpdaterString      = "NodeUpdater"
	edgeCreaterString      = "EdgeCreater"
	nodeDeleterString      = "NodeDeleter"
	edgeDeleterString      = "EdgeDeleter"
	nodeWriterString       = "NodeWriter"
	nodeAllerString        = "NodeAller"
	metaString             = "Meta"
	executorString         = "Executor"
	typerString            = "Typer"
	nodeTypedCreatorString = "NodeTypedCreator"
	nodeLabelerString      = "NodeLabeler"
	nodeLabelAdderString   = "NodeLabelAdder"
	edgeLabelerString      = "EdgeLabeler"
	edgeLabelAdderString   = "EdgeLabelAdder"
	closerString           = "Closer"
	dimensionChangerString = "DimensionChanger"
	graphCreaterString     = "GraphCreater"

	dimensionerString = "Dimensioner"
	dimensionsString  = "Dimensions"
	openReaderString  = "OpenReader"
	nodeOpenerString  = "NodeOpener"
	wraperString      = "Wraper"
)

func (c *command) dimensionChanger() (*walder.DimensionChanger, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.DimensionChanger)
	if !ok {
		return nil, fmt.Errorf("want walder.DimensionChanger, but got %T", g)
	}
	return &v, nil
}
func (c *command) dimensioner() (*walder.Dimensioner, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.Dimensioner)
	if !ok {
		return nil, fmt.Errorf("want walder.Dimensioner, but got %T", g)
	}
	return &v, nil
}
func (c *command) edgeCreater() (*walder.EdgeCreater, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.EdgeCreater)
	if !ok {
		return nil, fmt.Errorf("want walder.EdgeCreater, but got %T", g)
	}
	return &v, nil
}
func (c *command) executer() (*walder.Executor, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.Executor)
	if !ok {
		return nil, fmt.Errorf("want walder.Executor, but got %T", g)
	}
	return &v, nil
}
func (c *command) getReader() (*walder.GetReader, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.GetReader)
	if !ok {
		return nil, fmt.Errorf("want walder.GetReader, but got %T", g)
	}
	return &v, nil
}
func (c *command) graph() (*walder.Graph, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.Graph)
	if !ok {
		return nil, fmt.Errorf("want walder.Graph, but got %T", g)
	}
	return &v, nil
}
func (c *command) graphCreater() (*walder.GraphCreater, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.GraphCreater)
	if !ok {
		return nil, fmt.Errorf("want walder.GraphCreater, but got %T", g)
	}
	return &v, nil
}
func (c *command) directedGraph() (*walder.GraphDirected, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.GraphDirected)
	if !ok {
		return nil, fmt.Errorf("want walder.GraphDirected, but got %T", g)
	}
	return &v, nil
}
func (c *command) graphDirected() (*walder.GraphDirected, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.GraphDirected)
	if !ok {
		return nil, fmt.Errorf("want walder.GraphDirected, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeAller() (*walder.NodeAller, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeAller)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeAller, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeCreater() (*walder.NodeCreater, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeCreater)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeCreater, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeDeleter() (*walder.NodeDeleter, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeDeleter)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeDeleter, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeFromCreater() (*walder.NodeFromCreater, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeFromCreater)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeFromCreater, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeOpener() (*walder.NodeOpener, error) {
	n, err := c.node("")
	if err != nil {
		return nil, err
	}
	v, ok := n.(walder.NodeOpener)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeOpener, but got %T", n)
	}
	return &v, nil
}
func (c *command) nodeReader() (*walder.NodeReader, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeReader)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeReader, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeToCreater() (*walder.NodeToCreater, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeToCreater)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeToCreater, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeTypedCreator() (*walder.NodeTypedCreator, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeTypedCreator)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeTypedCreator, but got %T", g)
	}
	return &v, nil
}
func (c *command) nodeWriter() (*walder.NodeWriter, error) {
	g := c.walder.peek().graph
	v, ok := g.(walder.NodeWriter)
	if !ok {
		return nil, fmt.Errorf("want walder.NodeWriter, but got %T", g)
	}
	return &v, nil
}
func (c *command) openReader() (*walder.OpenReader, error) {
	n, err := c.walder.peek().getCursorItem()
	if err != nil {
		return nil, err
	}
	v, ok := n.(walder.OpenReader)
	if !ok {
		return nil, fmt.Errorf("want walder.OpenReader, but got %T", n)
	}
	return &v, nil
}

func (c *command) choose(options ...fmt.Stringer) (fmt.Stringer, error) {
	c.walder.Push(&chooser{
		from: options,
		execFunc: func(w *Walder, node fmt.Stringer) error {
			return nil
		}})
	c.pause("choose on of the nodes")

	// TODO check for correct chooser
	cur, err := c.walder.peek().getCursorItem()
	if err != nil {
		return nil, err
	}

	c.walder.pop()
	return cur, nil

}
func (c *command) node(string) (fmt.Stringer, error) {
	return c.walder.peek().getCursorItem()
}
func (c *command) repeat(string) (int, error) {
	if c.repeatAmount != "" {
		return strconv.Atoi(c.repeatAmount)
	}
	input, err := c.input("")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(input.String())
}
func (c *command) stack(lenght int) (stackTree, error) {
	if len(c.walder.stack.stack) < lenght {
		return stackTree{}, fmt.Errorf("stack is to short")
	}
	return c.walder.stack, nil
}
func (c *command) input(reason string) (fmt.Stringer, error) {
	var input string
	c.walder.activateInput(func(w *Walder, i string) error {
		input = i
		c.resume(w)

		c.walder.waitForCommand(c)

		return nil
	})

	c.pause(reason)

	return stringer(input), nil
}
func (c *command) graphHolder() (*graphHolder, error) {
	return c.walder.peek(), nil
}
func (c *command) holderList(editFunc func(l *holderList) error) error {
	focus := c.walder.focus()
	m := c.walder.peek().boxer.ModelMap[focus]
	l, ok := m.(holderList)
	if !ok {
		return fmt.Errorf("got %T, but want %T", m, l)
	}

	err := editFunc(&l)
	if err != nil {
		return err
	}

	c.walder.peek().boxer.ModelMap[focus] = l

	return nil
}
func (c *command) replaceNode(oldNode, newNode fmt.Stringer) error {
	return c.walder.peek().editList(mainAddr, func(l *holderList) error {
		i, err := l.GetCursorIndex()
		if err != nil {
			return err
		}
		return l.UpdateItem(i, func(old fmt.Stringer) (fmt.Stringer, error) {
			if oldNode != nil && old.String() != oldNode.String() {
				return nil, fmt.Errorf("did not found node to replace")
			}
			return newNode, nil
		})
	})
}
func (c *command) newNodes(nodes ...fmt.Stringer) {
	c.walder.addError(c.walder.peek().editList(mainAddr, func(l *holderList) error {
		return l.AddItems(nodes...)
	}))
	return
}
func (c *command) edgeList() ([][2]fmt.Stringer, error) {
	start, err := c.node("edge start")
	if err != nil {
		return nil, err
	}
	result, err := c.choose(stringer("in"), stringer("out"))
	if err != nil {
		return nil, err
	}
	b, err := c.graphHolder()
	if err != nil {
		return nil, err
	}
	switch result.String() {
	case "in":
		b.focus = inAddr
	case "out":
		b.focus = outAddr
	}
	c.pause("select edge list")
	end, err := c.nodeList("select edge targets")
	if err != nil {
		return nil, err
	}
	b, err = c.graphHolder()
	if err != nil {
		return nil, err
	}
	b.focus = mainAddr

	edges := make([][2]fmt.Stringer, 0, len(end))
	for _, e := range end {
		edges = append(edges, [2]fmt.Stringer{start, e})
	}

	return edges, nil
}
func (c *command) edgeDeleter() (*walder.EdgeDeleter, error) {
	g, err := c.graph()
	if err != nil {
		return nil, err
	}
	ed, ok := (*g).(walder.EdgeDeleter)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", ed, g)
	}
	return &ed, nil
}
func (c *command) nodeList(string) ([]fmt.Stringer, error) {
	return c.walder.peek().getCurrent()
}

func (c *command) returnGraph(g walder.Graph) {
	c.walder.addError(c.walder.Push(g))
	return
}
func editGraph[T walder.Graph](c *command, editFunc func(*T) error) error {
	g, err := c.graph()
	if err != nil {
		return err
	}
	t, ok := (*g).(T)
	if !ok {
		return fmt.Errorf("want %T, but got %T", t, *g)
	}
	return editFunc(&t)
}
