package lib

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	viz "github.com/awalterschulze/gographviz"
	"github.com/treilik/walder"
	"github.com/xtgo/uuid"
)

// DotDim is a generator for the Graph Viz Dimension
type DotDim struct{}

var _ walder.Dimensioner = DotDim{}

func (d DotDim) String() string {
	return "GraphViz"
}

// New returns a new GraphViz graph
func (d DotDim) New() (walder.Graph, error) {
	newGraph := &DotGraph{graph: viz.NewGraph()}
	newGraph.graph.Directed = true
	newGraph.graph.Attrs.Add("overlap", "scale")

	return newGraph, nil
}

var _ walder.OpenReader = DotDim{}

// Open satisfies the walder.OpenReader interface
func (d DotDim) Open(from io.Reader) (walder.Graph, error) {
	all, err := io.ReadAll(from)
	if err != nil {
		return nil, fmt.Errorf("error while opening from Reader: %w", err)
	}
	if len(all) == 0 {
		return &DotGraph{graph: viz.NewGraph()}, nil
	}

	ast, err := viz.Parse(all)
	if err != nil {
		return nil, fmt.Errorf("error while opening from Reader: %w", err)
	}

	graph, err := viz.NewAnalysedGraph(ast)
	if err != nil {
		return nil, fmt.Errorf("error while opening from Reader: %w", err)
	}

	return &DotGraph{graph: graph}, nil
}

type node viz.Node

func (n node) String() string {
	if l, ok := n.Attrs["label"]; ok {
		if len(l) > 2 &&
			strings.HasPrefix(l, "\"") &&
			strings.HasSuffix(l, "\"") &&
			!strings.HasSuffix(l, "\\\"") {

			l = l[1 : len(l)-1]
		}
		return l
	}
	return n.Name
}

type subgraph struct {
	graph viz.SubGraph
}

func (s subgraph) String() string {
	return s.graph.Name
}

func (n node) GetAttributes() [][2]string {
	attrList := make([][2]string, 0, len(n.Attrs))
	for k, v := range n.Attrs {
		attrList = append(attrList, [2]string{string(k), v})
	}

	return attrList
}

var _ walder.Graph = DotGraph{}
var _ walder.GraphDirected = DotGraph{}

type DotGraph struct {
	graph          *viz.Graph
	aktiveSubgraph string
}

func (d DotGraph) String() string {
	if d.graph == nil || d.graph.Name == "" {
		return "graphviz"
	}
	return fmt.Sprintf("graphviz: %s", d.graph.Name)
}

func (d DotGraph) HomeNodes() ([]fmt.Stringer, error) {
	return d.NodeAll()
}

func (d DotGraph) Outgoing(str fmt.Stringer) ([]fmt.Stringer, error) {
	n, ok := str.(node)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", n, str)
	}
	var nodeList []fmt.Stringer
	for e := range d.graph.Edges.SrcToDsts[n.Name] {
		out := node(*d.graph.Nodes.Lookup[e])
		if d.aktiveSubgraph != "" {
			// show only nodes of current subgraph
			parents, ok := d.graph.Relations.ChildToParents[out.Name]
			if ok && !parents[d.aktiveSubgraph] {
				continue
			}
		}

		nodeList = append(nodeList, out)
	}
	return sortStringer(nodeList), nil
}
func (d DotGraph) Incoming(str fmt.Stringer) ([]fmt.Stringer, error) {
	n, ok := str.(node)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", n, str)
	}
	var nodeList []fmt.Stringer
	for e := range d.graph.Edges.DstToSrcs[n.Name] {
		in := node(*d.graph.Nodes.Lookup[e])
		if d.aktiveSubgraph != "" {
			// show only nodes of current subgraph
			parents, ok := d.graph.Relations.ChildToParents[in.Name]
			if ok && !parents[d.aktiveSubgraph] {
				continue
			}
		}
		nodeList = append(nodeList, in)
	}

	return sortStringer(nodeList), nil
}

var _ walder.NodeCreater = &DotGraph{}

func (d *DotGraph) NodeCreate(input fmt.Stringer) (fmt.Stringer, error) {
	if input == nil {
		return nil, fmt.Errorf("recieved nil value")
	}
	in := escape(input.String())

	graph := d.graph.Name
	if d.aktiveSubgraph != "" {
		graph = d.aktiveSubgraph
	}

	id := quote(uuid.NewRandom().String())
	now := time.Now()
	created := fmt.Sprintf("created_%s_%s", now.Format(time.DateOnly), now.Format(time.TimeOnly))
	err := d.graph.AddNode(graph, id, map[string]string{"label": quote(in), "comment": created})
	if err != nil {
		return nil, err
	}

	n := d.graph.Nodes.Lookup[id]
	return node(*n), nil
}

func (d *DotGraph) NodeUpdate(toUpdate fmt.Stringer) (fmt.Stringer, error) {
	n, ok := toUpdate.(node)
	if !ok {
		return nil, fmt.Errorf("'%T' is not part of this graph", n)
	}
	existing, ok := d.graph.Nodes.Lookup[n.Name]
	if !ok {
		return nil, fmt.Errorf("'%s' is not part of this graph", n.String())
	}
	return node(*existing), nil
}

var _ walder.EdgeCreater = &DotGraph{}

func (d *DotGraph) EdgeCreate(from, to fmt.Stringer) error {
	// TODO allow supgraph as edge part when gographviz allows it
	fromNode, fromOK := from.(node)
	toNode, toOK := to.(node)

	if !fromOK || !toOK {
		return fmt.Errorf("can't create edge between '%T' and '%T'", from, to)
	}

	err := d.graph.AddEdge(fromNode.Name, toNode.Name, true, nil)
	if err == nil {
		d.graph.Directed = true
	}
	return err
}

var _ walder.NodeDeleter = &DotGraph{}

func (d *DotGraph) NodeDelete(str fmt.Stringer) error {
	node, ok := str.(node)
	if !ok {
		return fmt.Errorf("'%T' is not a node of this graph and thus can not be deleted", str)
	}

	err := d.graph.RemoveNode(d.graph.Name, node.Name)
	if err != nil {
		return err
	}

	return nil
}

var _ walder.GetReader = DotGraph{}

func (d DotGraph) GetReader() (io.Reader, error) {
	ast, err := d.graph.WriteAst()
	if err != nil {
		return nil, err
	}

	reader := bytes.NewBufferString(ast.String())
	return reader, nil
}

var _ walder.NodeToCreater = &DotGraph{}

func (d *DotGraph) NodeToCreate(input fmt.Stringer, toNodes ...fmt.Stringer) (fmt.Stringer, error) {
	if input == nil {
		return nil, fmt.Errorf("recieved nil value")
	}
	newNode, err := d.NodeCreate(input)
	if err != nil {
		return nil, err
	}

	for _, to := range toNodes {
		d.EdgeCreate(newNode, to)
	}
	return newNode, nil
}

var _ walder.NodeFromCreater = &DotGraph{}

func (d *DotGraph) NodeFromCreate(input fmt.Stringer, fromNodes ...fmt.Stringer) (fmt.Stringer, error) {
	if input == nil {
		return nil, fmt.Errorf("recieved nil value")
	}
	newNode, err := d.NodeCreate(input)
	if err != nil {
		return nil, err
	}

	for _, from := range fromNodes {
		d.EdgeCreate(from, newNode)
	}
	return newNode, nil
}

var _ walder.NodeAller = DotGraph{}

func (d DotGraph) NodeAll() ([]fmt.Stringer, error) {
	if d.graph.Nodes == nil {
		d = DotGraph{graph: viz.NewGraph()}
	}
	var start []fmt.Stringer
	for _, n := range d.graph.Nodes.Nodes {
		if d.aktiveSubgraph != "" {
			// show only nodes of current subgraph
			parents, ok := d.graph.Relations.ChildToParents[n.Name]
			if ok && !parents[d.aktiveSubgraph] {
				continue
			}
		}
		start = append(start, node(*n))
	}
	for _, s := range d.graph.SubGraphs.SubGraphs {
		if d.aktiveSubgraph != "" {
			// show only nodes of current subgraph
			parents, ok := d.graph.Relations.ChildToParents[s.Name]
			if ok && !parents[d.aktiveSubgraph] {
				continue
			}
		}
		start = append(start, subgraph{*s})
	}
	return start, nil
}

var _ walder.Typer = &DotGraph{}

func (d *DotGraph) GetType(n fmt.Stringer) (string, error) {
	switch n.(type) {
	case node:
		return "node", nil
	case subgraph:
		return "subgraph", nil
	}
	return "", fmt.Errorf("unknown type %#v", n)
}

var _ walder.NodeTypedCreator = &DotGraph{}

var types []fmt.Stringer = []fmt.Stringer{
	stringer("node"),
	stringer("subgraph"),
}

func (d *DotGraph) GetTypes() ([]fmt.Stringer, error) {
	return types, nil
}

func (d *DotGraph) NodeTypedCreate(Type fmt.Stringer, input fmt.Stringer) (fmt.Stringer, error) {
	i := input.String()
	if Type.String() == "node" {
		d.graph.AddNode(d.aktiveSubgraph, i, nil)
		n := d.graph.Nodes.Lookup[i]
		return node(*n), nil
	}
	if Type.String() == "subgraph" {
		d.graph.AddSubGraph(d.aktiveSubgraph, i, nil)
		s := d.graph.SubGraphs.SubGraphs[i]
		return subgraph{*s}, nil
	}
	return nil, fmt.Errorf("no type named '%s'", Type)
}

var _ walder.NodeOpener = &DotGraph{}

func (d *DotGraph) NodeOpen(nodes ...fmt.Stringer) (walder.Graph, error) {
	if len(nodes) != 1 {
		return nil, fmt.Errorf("can only open one node, nut %d", len(nodes))
	}
	node := nodes[0]
	s, ok := node.(subgraph)
	if !ok {
		return nil, fmt.Errorf("cant enter '%v'", node)
	}
	d.aktiveSubgraph = s.graph.Name
	return d, nil
}

var _ walder.NodeLabelAdder = DotGraph{}

func (d DotGraph) NodeLabels(n fmt.Stringer) ([][2]string, error) {
	nd, ok := n.(node)
	labels := make([][2]string, 0, len(d.graph.Nodes.Lookup[nd.Name].Attrs))
	if ok {
		for k, v := range d.graph.Nodes.Lookup[nd.Name].Attrs { // TODO fix panic when node not found
			labels = append(labels, [2]string{string(k), v})
		}
		return labels, nil
	}
	sg, ok := n.(subgraph)
	if ok {
		for k, v := range d.graph.SubGraphs.SubGraphs[sg.graph.Name].Attrs { // TODO fix panic when node not found
			labels = append(labels, [2]string{string(k), v})
		}
		return labels, nil
	}

	return nil, fmt.Errorf("unhandled type: %T", n)
}
func (d DotGraph) NodeLabelAdd(n fmt.Stringer, key, value string) (fmt.Stringer, error) {
	nd, ok := n.(node)
	if ok {
		d.graph.Nodes.Lookup[nd.Name].Attrs.Add(key, value)
		n, ok := d.graph.Nodes.Lookup[nd.Name]
		if !ok {
			return nil, fmt.Errorf("node not found")
		}
		return node(*n), nil

	}
	sg, ok := n.(subgraph)
	if ok {
		d.graph.SubGraphs.SubGraphs[sg.graph.Name].Attrs.Add(key, value)
		sg, ok := d.graph.SubGraphs.SubGraphs[sg.graph.Name]
		if !ok {
			return nil, fmt.Errorf("subgraph not found")
		}
		return subgraph{*sg}, nil
	}
	return nil, fmt.Errorf("unhandled type: %T", n)
}

var _ walder.EdgeDeleter = &DotGraph{}

func (d *DotGraph) EdgeDelete(from, to fmt.Stringer) error {
	fromNode, fromOK := from.(node)
	toNode, toOK := to.(node)
	if !fromOK || !toOK {
		return fmt.Errorf("can only operate on internal nodes")
	}

	for c := 0; c < len(d.graph.Edges.Edges); c++ {
		e := d.graph.Edges.Edges[c]
		if e.Src != fromNode.Name || e.Dst != toNode.Name {
			continue
		}
		{
			lenght := len(d.graph.Edges.Edges)
			var rest []*viz.Edge
			if lenght > c+1 {
				rest = d.graph.Edges.Edges[c+1:]
			}
			d.graph.Edges.Edges = append(d.graph.Edges.Edges[:c], rest...)
			delete(d.graph.Edges.SrcToDsts[fromNode.Name], toNode.Name)
			delete(d.graph.Edges.DstToSrcs[toNode.Name], fromNode.Name)
		}
	}
	return nil
}

var _ walder.EdgeLabeler = DotGraph{}

func (d DotGraph) EdgeLabels(from, to fmt.Stringer) ([][2]string, error) {
	f, fOK := from.(node)
	t, tOK := to.(node)
	if !fOK || !tOK {
		return nil, fmt.Errorf("want %T but have %T and %T", f, from, to)
	}
	dest, ok := d.graph.Edges.SrcToDsts[f.Name]
	if !ok {
		return nil, fmt.Errorf("node not found %s", f.Name)
	}
	e, ok := dest[t.Name]
	if !ok {
		return nil, fmt.Errorf("node not found %s", f.Name)
	}
	if e == nil {
		return nil, fmt.Errorf("no edge found")
	}
	if len(e) != 1 {
		return nil, fmt.Errorf("only able to handle exactly one edge per node pair not %d", len(e))
	}
	attrs := e[0].Attrs
	labels := make([][2]string, 0, len(attrs))
	for k, v := range attrs {
		labels = append(labels, [2]string{string(k), v})
	}

	return labels, nil
}

var _ walder.EdgeLabelAdder = DotGraph{}

func (d DotGraph) EdgeLabelAdd(from, to fmt.Stringer, k, v string) error {
	f, fOK := from.(node)
	t, tOK := to.(node)
	if !fOK || !tOK {
		return fmt.Errorf("want %T but have %T and %T", f, from, to)
	}
	dest, ok := d.graph.Edges.SrcToDsts[f.Name]
	if !ok {
		return fmt.Errorf("node not found %s", f.Name)
	}
	e, ok := dest[t.Name]
	if !ok {
		return fmt.Errorf("node not found %s", f.Name)
	}
	if e == nil {
		return fmt.Errorf("no edge found")
	}
	if len(e) != 1 {
		return fmt.Errorf("only able to handle exactly one edge per node pair not %d", len(e))
	}
	newEdge := e[0]
	newEdge.Attrs.Add(k, v)
	for i, e := range d.graph.Edges.Edges {
		if e.Src != f.Name || e.Dst != t.Name {
			continue
		}
		d.graph.Edges.Edges[i] = newEdge
	}
	d.graph.Edges.SrcToDsts[f.Name][t.Name][0] = newEdge
	d.graph.Edges.DstToSrcs[t.Name][f.Name][0] = newEdge
	return nil
}

var _ walder.Closer = &DotGraph{}

func (d *DotGraph) Close() error {
	// TODO handle nested subgraph returns

	if d.aktiveSubgraph == "" {
		return fmt.Errorf("cant leave base")
	}
	d.aktiveSubgraph = ""
	return nil

}

var _ walder.NodeSwaper = &DotGraph{}

func (d *DotGraph) NodeSwap(first, second fmt.Stringer) error {
	f, ok := first.(node)
	if !ok {
		return fmt.Errorf("want %T but have %T ", f, first)
	}
	s, ok := second.(node)
	if !ok {
		return fmt.Errorf("want %T but have %T ", s, second)
	}

	edges := viz.NewEdges()

	for _, edge := range d.graph.Edges.Edges {
		newSrc := edge.Src
		newDst := edge.Dst
		if edge.Src == f.Name {
			newSrc = s.Name
		}
		if edge.Src == s.Name {
			newSrc = f.Name
		}
		if edge.Dst == f.Name {
			newDst = s.Name
		}
		if edge.Dst == s.Name {
			newDst = f.Name
		}
		edge.Src = newSrc
		edge.Dst = newDst
		edges.Add(edge)
	}
	d.graph.Edges = edges
	return nil
}

func escape(in string) string {
	return strings.ReplaceAll(in, "\"", "\\\"")
}

func quote(in string) string {
	return fmt.Sprintf("\"%s\"", in)
}
