package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dominikbraun/graph"
	"github.com/treilik/walder"
)

type cmdDag struct {
	bindings graph.Graph[int, keyBinding]

	mu        *sync.Mutex
	idCounter int
}

var _ internal = &cmdDag{}

func (c *cmdDag) renew(w *Walder) {
	// TODO document
	w.bindings = c
	return
}

var _ walder.OpenReader = &cmdDag{}

func (c *cmdDag) Open(r io.Reader) (walder.Graph, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return c, err
	}
	newBindings := make(map[string]string)
	err = json.Unmarshal(content, &newBindings)
	if err != nil {
		return c, err
	}

	if c.bindings == nil {
		c.newGraph()
	}

	internals := make(map[string]keyBinding)

	for _, cmd := range commandList {
		b := keyBinding{Cmd: cmd, id: c.nextID()}
		err := c.bindings.AddVertex(b)
		if err != nil {
			return c, fmt.Errorf("while adding commands an error occured: %w", err)
		}
		for k, v := range newBindings {
			if cmd.Name != k {
				continue
			}

			keyOrder := []tea.KeyMsg{}

			keys := strings.Split(v, ",") // TODO dont hardcode and make , bindabel
			var former keyBinding
			for _, k := range keys {
				key, err := getKey(k)
				if err != nil {
					return c, err
				}
				keyOrder = append(keyOrder, tea.KeyMsg(key))
				var n keyBinding
				var ok bool
				if n, ok = internals[keyBinding{KeyOrder: keyOrder}.keyOrderString()]; !ok {
					n = keyBinding{KeyOrder: keyOrder, id: c.nextID()}
					if err := c.bindings.AddVertex(n); err != nil {
						return c, err
					}
				}
				if former.id > 0 {
					// form path
					if err := c.bindings.AddEdge(former.id, n.id); err != nil {
						return c, err
					}
				}
				former = n
			}
			b.KeyOrder = keyOrder
			// Update Node
			if err := c.bindings.AddVertex(b); err != nil {
				return c, err
			}
			// complete path
			if former.id > 0 {
				if err := c.bindings.AddEdge(former.id, b.id); err != nil {
					return c, err
				}
			}
			break
		}
	}
	return c, nil
}

var _ walder.GetReader = cmdDag{}

func (c cmdDag) GetReader() (io.Reader, error) {
	bindings := make(map[string]string)
	adjMap, err := c.bindings.AdjacencyMap()
	if err != nil {
		return nil, err
	}

	for id, _ := range adjMap {
		v, err := c.bindings.Vertex(id)
		if v.Cmd.run == nil {
			continue
		}
		if len(v.KeyOrder) == 0 {
			continue
		}
		if err != nil {
			return nil, err
		}
		bindings[v.Cmd.Name] = v.keyOrderString()
	}
	by, err := json.MarshalIndent(&bindings, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(by), nil
}

var _ walder.Graph = cmdDag{}

func (c cmdDag) String() string {
	return "bindings"
}

type keyBinding struct {
	KeyOrder []tea.KeyMsg
	Cmd      command
	id       int
}

func (b keyBinding) keyOrderString() string {
	bString := make([]string, 0, len(b.KeyOrder))
	for _, k := range b.KeyOrder {
		bString = append(bString, k.String())
	}
	return strings.Join(bString, ",")
}

var _ walder.Graph = cmdDag{}

func (b keyBinding) String() string {
	if b.Cmd.run != nil {
		if b.Cmd.Description != "" {
			return fmt.Sprintf("%s\n%s", b.Cmd.Name, b.Cmd.Description)
		}
		return b.Cmd.Name
	}
	return b.KeyOrder[len(b.KeyOrder)-1].String()
}

func (c *cmdDag) Add(cmdList []command) error {
	if c.bindings == nil {
		c.newGraph()
	}
	for _, cmd := range cmdList {
		newID := c.nextID()
		b := keyBinding{
			id:  newID,
			Cmd: cmd,
		}
		err := c.bindings.AddVertex(b)
		if err != nil {
			return err
		}
	}
	return nil
}

// HomeNodes returns all nodes of this graph
func (c cmdDag) HomeNodes() ([]fmt.Stringer, error) {
	stringerList := make([]fmt.Stringer, 0, c.bindings.Order())
	adjMap, err := c.bindings.AdjacencyMap()
	if err != nil {
		return nil, err
	}
	for id, _ := range adjMap {
		v, err := c.bindings.Vertex(id)
		if err != nil {
			return sortStringer(stringerList), err
		}
		if v.Cmd.run == nil {
			continue
		}
		stringerList = append(stringerList, v)
	}

	return sortStringer(stringerList), nil
}

var _ walder.GraphDirected = cmdDag{}

func (c cmdDag) Incoming(str fmt.Stringer) ([]fmt.Stringer, error) {
	node, ok := str.(keyBinding)
	if !ok {
		return nil, fmt.Errorf("not a keyBinding")
	}
	preMap, err := c.bindings.PredecessorMap()
	if err != nil {
		return nil, err
	}
	var stringerList []fmt.Stringer
	for id, _ := range preMap[node.id] {
		v, err := c.bindings.Vertex(id)
		if err != nil {
			return stringerList, err
		}
		stringerList = append(stringerList, v)
	}
	return stringerList, nil

}

func (c cmdDag) Outgoing(str fmt.Stringer) ([]fmt.Stringer, error) {
	node, ok := str.(keyBinding)
	if !ok {
		return nil, fmt.Errorf("not a keyBinding")
	}
	adjMap, err := c.bindings.AdjacencyMap()
	if err != nil {
		return nil, err
	}
	var stringerList []fmt.Stringer
	for id, _ := range adjMap[node.id] {
		v, err := c.bindings.Vertex(id)
		if err != nil {
			return stringerList, err
		}
		stringerList = append(stringerList, v)
	}
	return stringerList, nil
}

func (c cmdDag) Filter(path []tea.KeyMsg) ([]keyBinding, error) {
	if len(path) == 0 {
		bindings := make([]keyBinding, 0, c.bindings.Order())
		nodes, err := c.nodes()
		if err != nil {
			return nil, err
		}
		for _, v := range nodes {
			// only include leafs, so exclude internal nodes
			if v.Cmd.run == nil {
				continue
			}
			bindings = append(bindings, v)
		}
		return bindings, nil
	}

	var reachableCmds []keyBinding

	nodes, err := c.nodes()
	if err != nil {
		return nil, err
	}
	for _, b := range nodes {
		// only include leafs, so exclude internal nodes
		if b.Cmd.run == nil {
			continue
		}
		if len(b.KeyOrder) < len(path) {
			continue
		}
		if strings.HasPrefix(b.keyOrderString(), keyBinding{KeyOrder: path}.keyOrderString()) {
			reachableCmds = append(reachableCmds, b)
		}
	}
	return reachableCmds, nil
}

var _ walder.NodeCreater = &cmdDag{}

func (c *cmdDag) NodeCreate(input fmt.Stringer) (fmt.Stringer, error) {
	if input == nil {
		return nil, fmt.Errorf("recived nil value")
	}
	key, err := getKey(input.String())
	if err != nil {
		return nil, err
	}
	b := keyBinding{
		KeyOrder: []tea.KeyMsg{tea.KeyMsg(key)},
		id:       c.nextID(),
	}
	err = c.bindings.AddVertex(b)
	return b, err
}

var _ walder.EdgeCreater = &cmdDag{}

func (c *cmdDag) NodeUpdate(node fmt.Stringer) (fmt.Stringer, error) {
	n, ok := node.(keyBinding)
	if !ok {
		return nil, fmt.Errorf("'%T' is not part of this graph", node)
	}

	if n.Cmd.run == nil {
		return node, nil
	}

	kb, err := c.bindings.Vertex(n.id)
	return kb, err
}

func (c *cmdDag) EdgeCreate(from, to fmt.Stringer) error {
	fromNode, fromOk := from.(keyBinding)
	toNode, toOk := to.(keyBinding)
	if !fromOk || !toOk {
		return fmt.Errorf("can only create edge between keyBinding, not from '%T' to '%T'", from, to)
	}

	if fromNode.Cmd.run != nil {
		return fmt.Errorf("commands can only be leaves, thus no edge can originate from them")
	}

	// check if nodes are of this graph
	if _, err := c.bindings.Vertex(toNode.id); err != nil {
		return err
	}
	if _, err := c.bindings.Vertex(fromNode.id); err != nil {
		return err
	}

	// check if from node is invalid
	if len(fromNode.KeyOrder) == 0 {
		return fmt.Errorf("from node must containe a key but has none")
	}

	oldPrefix := toNode.keyOrderString()
	if toNode.Cmd.run != nil {
		// prepend keys
		toNode.KeyOrder = append(fromNode.KeyOrder, toNode.KeyOrder...)

		// update intern
		if err := c.bindings.AddVertex(toNode); err != nil { // TODO remove node and then add.?
			return err
		}
		return c.bindings.AddEdge(fromNode.id, toNode.id)
	}
	adjMap, err := c.bindings.AdjacencyMap() // TODO ask only for nodes
	if err != nil {
		return err
	}
	for id, _ := range adjMap {
		cmd, err := c.bindings.Vertex(id)
		if err != nil {
			return err
		}
		if strings.HasPrefix(cmd.keyOrderString(), oldPrefix) {
			cmd.KeyOrder = append(fromNode.KeyOrder, cmd.KeyOrder...)
			err := c.bindings.AddVertex(cmd) // TODO remove node and then add.?
			if err != nil {
				return err
			}
		}
	}
	return c.bindings.AddEdge(fromNode.id, toNode.id)
}

func (c *cmdDag) nextID() int {
	if c.mu == nil {
		c.mu = &sync.Mutex{}
	}
	c.mu.Lock()
	c.idCounter++
	newID := c.idCounter
	c.mu.Unlock()

	return newID
}

func (c *cmdDag) newGraph() {
	c.bindings = graph.New(func(kb keyBinding) int { return kb.id }, graph.Acyclic(), graph.Directed())
}

func (c *cmdDag) nodes() ([]keyBinding, error) {
	adjMap, err := c.bindings.AdjacencyMap()
	if err != nil {
		return nil, err
	}
	var nodes []keyBinding
	for id, _ := range adjMap {
		v, err := c.bindings.Vertex(id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, v)
	}
	return nodes, nil
}

var _ walder.EdgeDeleter = &cmdDag{}

func (c *cmdDag) EdgeDelete(from, to fmt.Stringer) error {
	err := c.edgeDelete(from, to)
	return err
}
func (c *cmdDag) edgeDelete(from, to fmt.Stringer) error {
	f, fOK := from.(keyBinding)
	t, tOK := to.(keyBinding)
	if !fOK || !tOK {
		return fmt.Errorf("%T and %T have to be %T", from, to, keyBinding{})
	}
	if err := c.bindings.RemoveEdge(f.id, t.id); err != nil {
		return err
	}

	v, err := c.bindings.Vertex(t.id)
	if err != nil {
		return err
	}
	// TODO remove vertex
	v.KeyOrder = []tea.KeyMsg{}
	if err := c.bindings.AddVertex(v); err != nil {
		return err
	}

	// recursivley delete edges
	edges, err := c.Outgoing(v)
	if err != nil {
		return err
	}
	for _, e := range edges {
		err := c.edgeDelete(v, e)
		if err != nil {
			return err
		}
	}
	return nil
}
