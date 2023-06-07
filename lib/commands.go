package lib

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/treilik/walder"
)

var (
	commandList = []command{
		{
			Name:        "remove from to",
			Description: "",
			run: func(c *command) error {
				from, err := c.node("from")
				if err != nil {
					return err
				}
				c.pause("get 'to' node")
				to, err := c.node("to")
				if err != nil {
					return err
				}
				toString := to.String()
				gd, err := c.graphDirected()
				if err != nil {
					return err
				}

				var paths [][]fmt.Stringer
				err = dfs(*gd,
					func(path []fmt.Stringer) (bool, error) {
						if len(path) == 0 {
							return false, nil
						}
						node := path[len(path)-1]
						for _, n := range path[:len(path)-1] {
							if n.String() == node.String() {
								abort := true
								return abort, nil
							}
						}
						if node.String() == toString {
							paths = append(paths, path)
						}
						return false, nil
					},
					from,
				)
				if err != nil {
					return err
				}

				c.pause("get graph where the limited graph should be writen to")

				gc, err := c.graphCreater()
				if err != nil {
					return err
				}

				for _, path := range paths {
					var former fmt.Stringer
					for i, n := range path {
						if existingNode, err := (*gc).NodeUpdate(n); err == nil {
							n = existingNode
						} else {
							n, err = (*gc).NodeCreate(n)
							if err != nil {
								return err
							}
						}

						if i > 0 {
							err := (*gc).EdgeCreate(former, n)
							if err != nil {
								return err
							}
						}
						former = n
					}
				}

				c.returnGraph(*gc)
				return err
			},
		},
		{
			Name:        "rename node",
			Description: "",
			run: func(c *command) error {
				node, err := c.node("")
				if err != nil {
					return err
				}
				g, err := c.graph()
				if err != nil {
					return err
				}
				nn, ok := (*g).(walder.NodeNamer)
				if !ok {
					return fmt.Errorf("want %T, but got %T", nn, *g)
				}
				input, err := c.input("")
				if err != nil {
					return err
				}
				New, err := nn.NodeName(node, input.String())
				if err != nil {
					return err
				}
				return c.replaceNode(nil, New)
			},
		},
		{
			Name:        "break cycles",
			Description: "",
			run: func(c *command) error {
				node, err := c.node("")
				if err != nil {
					return err
				}
				gd, err := c.graphDirected()
				if err != nil {
					return err
				}
				c.pause("choose where to write graph with out cycles")
				gc, err := c.graphCreater()
				if err != nil {
					return err
				}
				return BreakCycles(node, *gd, *gc)
			},
		},
		{
			Name:        "edge move",
			Description: "",
			run: func(c *command) error {
				toMove, err := c.node("to move")
				c.pause("from")
				from, err := c.node("from")
				if err != nil {
					return err
				}

				c.pause("to")
				to, err := c.node("to")
				if err != nil {
					return err
				}
				return editGraph(c, func(swaper *walder.EdgeMover) error {
					return (*swaper).EdgeMove(toMove, from, to)
				})
			},
		},
		{
			Name:        "node swap",
			Description: "",
			run: func(c *command) error {
				first, err := c.node("first")
				if err != nil {
					return err
				}
				c.pause("to get the node with which to swap")
				second, err := c.node("second")
				if err != nil {
					return err
				}
				return editGraph(c, func(swaper *walder.NodeSwaper) error {
					return (*swaper).NodeSwap(first, second)
				})
			},
		},
		{
			Name:        "stack",
			Description: "",
			run: func(c *command) error {
				c.returnGraph(c.walder.stack)
				return nil
			},
		},
		{
			Name:        "ball",
			Description: "",
			run: func(c *command) error {
				_, err := c.graphDirected()
				if err != nil {
					return err
				}
				size, err := c.repeat("size")
				if err != nil {
					return err
				}
				gd, err := c.graphDirected()
				if err != nil {
					return err
				}
				current, err := c.node("from")
				if err != nil {
					return err
				}
				c.returnGraph(ball{
					super:  *gd,
					origin: current,
					size:   size,
				})
				return nil
			},
		},
		{
			Name:        "transitive reduce",
			Description: "",
			run: func(c *command) error {
				g, err := c.graph()
				if err != nil {
					return err
				}
				fd, ok := (*g).(finiteDeleter)
				if !ok {
					return fmt.Errorf("want %T, but got %T", fd, *g)
				}
				return reduction(fd)
			},
		},
		{
			Name:        "delete edge",
			Description: "",
			run: func(c *command) error {
				_, err := c.edgeDeleter()
				if err != nil {
					return err
				}
				edges, err := c.edgeList()
				if err != nil {
					return err
				}
				ed, err := c.edgeDeleter()
				if err != nil {
					return err
				}
				for _, e := range edges {
					(*ed).EdgeDelete(e[0], e[1])
				}
				return nil
			},
		},
		{
			Name:        "topo sort",
			Description: "",
			run: func(c *command) error {
				gd, err := c.graphDirected()
				if err != nil {
					return err
				}
				sorted, err := topo{
					origin: *gd,
				}.TopoSort()
				return c.walder.peek().editList(mainAddr, func(l *holderList) error {
					return l.ResetItems(sorted...)
				})
			},
		},
		{
			Name:        "save main list",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					all, err := l.GetAllItems()
					if err != nil {
						return err
					}
					b := &bytes.Buffer{}
					for _, a := range all {
						b.WriteString(a.String())
						b.WriteString("\n")
					}
					c.pause("get graph and node to which to write")
					nw, err := c.nodeWriter()
					if err != nil {
						return err
					}
					n, err := c.node("node to write to")
					if err != nil {
						return err
					}
					w, err := (*nw).NodeWrite(n)
					if err != nil {
						return err
					}
					_, err = io.Copy(w, b)
					return err
				})
			},
		},
		{
			Name:        "beginning",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {

					return l.Top()
				})
			},
		},
		{
			Name:        "end",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					return l.Bottom()
				})
			},
		},
		{
			Name:        "create incoming node",
			Description: "",
			run: func(c *command) error {
				transferNode, err := c.input("new name")
				if err != nil {
					return err
				}
				existingNodes, err := c.nodeList("descendants of the new node")
				if err != nil {
					return err
				}

				ntc, err := c.nodeToCreater()
				if err == nil {
					_, err = (*ntc).NodeToCreate(transferNode, existingNodes...)
					return err
				}

				nc, err := c.nodeCreater()
				if err != nil {
					return err
				}
				ec, err := c.edgeCreater()
				if err != nil {
					return err
				}
				newNode, err := (*nc).NodeCreate(transferNode)
				if err != nil {
					return err
				}
				for _, s := range existingNodes {
					err := (*ec).EdgeCreate(newNode, s)
					if err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name:        "create outgoing node",
			Description: "",
			run: func(c *command) error {
				transferNode, err := c.input("new name")
				if err != nil {
					return err
				}
				existingNodes, err := c.nodeList("ancestors of the new node")
				if err != nil {
					return err
				}

				nfc, err := c.nodeFromCreater()
				if err == nil {
					_, err = (*nfc).NodeFromCreate(transferNode, existingNodes...)
					return err
				}

				nc, err := c.nodeCreater()
				if err != nil {
					return err
				}
				ec, err := c.edgeCreater()
				if err != nil {
					return err
				}
				newNode, err := (*nc).NodeCreate(transferNode)
				if err != nil {
					return err
				}
				for _, s := range existingNodes {
					err := (*ec).EdgeCreate(s, newNode)
					if err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name:        "toggle selected",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					i, err := l.GetCursorIndex()
					if err != nil {
						return err
					}
					j, err := c.repeat("jump")
					if err != nil {
						j = 1
					}

					d := 1
					if j < 0 {
						d = -1
					}
					for c := 0; c < j; c++ {
						index := i + (c * d)

						value, err := l.GetSelect(index)
						if err != nil {
							return err
						}
						if err := l.SetSelect(index, !value); err != nil {
							return err
						}
					}
					_, err = l.SetCursor(i + (j * d))
					return err
				})
			},
		},
		{
			Name:        "left",
			Description: "",
			run: func(c *command) error {
				b, err := c.graphHolder()
				if err != nil {
					return err
				}
				return b.moveIncoming()
			},
		},
		{
			Name:        "right",
			Description: "",
			run: func(c *command) error {
				b, err := c.graphHolder()
				if err != nil {
					return err
				}
				return b.moveOutgoing()
			},
		},
		{
			Name:        "up",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					_, err := l.MoveCursor(-1)
					return err
				})
			},
		},
		{
			Name:        "down",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					_, err := l.MoveCursor(1)
					return err
				})
			},
		},
		{
			Name:        "enter in new graph",
			Description: "",
			run: func(c *command) error {
				cur, err := c.node("node to enter")
				if err != nil {
					return err
				}
				graph, ok := cur.(walder.Graph)
				if !ok {
					return fmt.Errorf("want %s, but got %T", graphString, cur)
				}
				c.returnGraph(graph)
				return nil
			},
		},
		{
			Name:        "enter in new Dimension",
			Description: "",
			run: func(c *command) error {
				cur, err := c.node("node to enter")
				if err != nil {
					return err
				}
				dim, ok := cur.(walder.Dimensioner)
				if !ok {
					return fmt.Errorf("want %s, but got %T", dimensionerString, cur)
				}

				newGraph, err := dim.New()
				if err != nil {
					return err
				}
				c.returnGraph(newGraph)
				return nil
			},
		},
		{
			Name:        "back",
			Description: "go back to the graph before current one.\nthis does not save any work or changes you made to the current one",
			run: func(c *command) error {
				c.walder.pop()
				return nil
			},
		},
		{
			Name:        "create new node",
			Description: "create new node in this graph from input text",
			run: func(c *command) error {
				nc, err := c.nodeCreater()
				if err != nil {
					return err
				}
				input, err := c.input("new name")
				if err != nil {
					return err
				}
				newNode, err := (*nc).NodeCreate(input)
				if err != nil {
					return err
				}
				c.newNodes(newNode)
				return nil
			},
		},
		{
			Name:        "start new edge",
			Description: "start creating a directed edge from or to the current selection of nodes",
			run: func(c *command) error {
				ec, err := c.edgeCreater()
				if err != nil {
					return err
				}

				first, err := c.nodeList("first")
				if err != nil {
					return err
				}
				c.pause("get second nodes")

				n, err := c.choose(stringer("to"), stringer("from"))
				if err != nil {
					return err
				}
				direction := n.String()

				second, err := c.nodeList("second")
				if err != nil {
					return err
				}

				for _, f := range first {
					for _, s := range second {
						switch direction {
						case "to":
							(*ec).EdgeCreate(f, s)
						case "from":
							(*ec).EdgeCreate(s, f)
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "open from reader",
			Description: "open a new graph adapter from the current node",
			run: func(c *command) error {
				cur, err := c.node("to open")
				if err != nil {
					return err
				}

				nr, err := c.nodeReader()
				if err != nil {
					return err
				}
				reader, err := (*nr).NodeRead(cur)
				if err != nil {
					return err
				}

				c.pause("get dimension in which to open read node")

				or, err := c.openReader()
				if err != nil {
					return err
				}
				newGraph, err := (*or).Open(reader)
				if err != nil {
					return err
				}
				c.returnGraph(newGraph)
				return nil
			},
		},

		{
			Name:        "write graph to former node",
			Description: "steay in this graph and try to write its content to current node of the former graph",
			run: func(c *command) error {
				s, err := c.stack(2)
				if err != nil {
					return nil
				}
				gr, err := c.getReader()
				if err != nil {
					return err
				}
				r, err := (*gr).GetReader()
				if err != nil {
					return err
				}

				top, err := s.pop()

				top.update()
				defer func() { s.push(top) }()

				former, err := s.peek()
				if err != nil {
					return err
				}

				var cur fmt.Stringer
				var formerNodeIndex int
				err = former.editList(mainAddr, func(l *holderList) error {
					formerNodeIndex, _ = l.GetCursorIndex()
					cur, err = l.GetCursorItem()
					return err
				})
				if err != nil {
					return fmt.Errorf("error while retriving curser node from former graph: %w", err)
				}
				nw, err := c.nodeWriter()
				if err != nil {
					return err
				}

				wc, err := (*nw).NodeWrite(cur)
				if err != nil {
					return err
				}
				if _, err := io.Copy(wc, r); err != nil {
					return err
				}
				wc.Close()
				n, err := (*nw).NodeUpdate(cur)
				if err != nil {
					return err
				}
				return former.editList(mainAddr, func(l *holderList) error {
					return l.UpdateItem(formerNodeIndex, func(_ fmt.Stringer) (fmt.Stringer, error) {
						return n, nil
					})
				})

			},
		},
		{
			Name:        "leave and write graph",
			Description: "leave this graph and try to write its content to current node of the former graph",
			run: func(c *command) error {
				gr, err := c.getReader()
				if err != nil {
					return err
				}

				s, err := c.stack(2)
				if err != nil {
					return err
				}

				s.pop()

				n, err := c.node("write to")
				if err != nil {
					return err
				}

				nw, err := c.nodeWriter()
				if err != nil {
					return err
				}
				r, err := (*gr).GetReader()
				if err != nil {
					return err
				}
				wr, err := (*nw).NodeWrite(n)
				if err != nil {
					return err
				}

				_, err = io.Copy(wr, r)
				if err != nil {
					return err
				}

				newNode, err := (*nw).NodeUpdate(n)
				if err != nil {
					return err
				}

				c.replaceNode(n, newNode)
				return nil
			},
		},
		{
			Name:        "write graph",
			Description: "",
			run: func(c *command) error {
				gr, err := c.getReader()
				if err != nil {
					return err
				}
				r, err := (*gr).GetReader()
				if err != nil {
					return err
				}

				c.pause("get node and graph where to write to")
				nw, err := c.nodeWriter()
				if err != nil {
					return err
				}
				n, err := c.node("write to")
				if err != nil {
					return err
				}
				wr, err := (*nw).NodeWrite(n)
				if err != nil {
					return err
				}

				_, err = io.Copy(wr, r)
				if err != nil {
					return err
				}
				err = wr.Close()
				if err != nil {
					return err
				}

				newNode, err := (*nw).NodeUpdate(n)
				if err != nil {
					return err
				}

				c.replaceNode(n, newNode)
				return nil
			},
		},
		{
			Name:        "home",
			Description: "",
			run: func(c *command) error {
				g, err := c.graph()
				if err != nil {
					return err
				}
				home, err := (*g).HomeNodes()
				if err != nil {
					return err
				}
				return c.holderList(func(l *holderList) error {
					return l.ResetItems(home...)
				})
			},
		},
		{
			Name:        "get all nodes",
			Description: "",
			run: func(c *command) error {
				na, err := c.nodeAller()
				if err != nil {
					return err
				}
				all, err := (*na).NodeAll()
				if err != nil {
					return err
				}
				return c.holderList(func(l *holderList) error {
					return l.ResetItems(all...)
				})
			},
		},
		{
			Name:        "sparse modus",
			Description: "",
			run: func(c *command) error {
				cur, err := c.walder.peek().getCursorItem()
				if err != nil {
					return err
				}
				g := c.walder.pop()
				t := newSparseModus(g.graph)
				t.resetMain(cur)
				return c.walder.Push(g.graph, t)
			},
		},
		{
			Name:        "tree modus",
			Description: "",
			run: func(c *command) error {
				cur, err := c.walder.peek().getCursorItem()
				if err != nil {
					return err
				}
				g := c.walder.pop()
				t := newTreeModus(g.graph)
				t.resetMain(cur)
				return c.walder.Push(g.graph, t)
			},
		},
		{
			Name:        "string sort",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					sort.Sort(l)
					return nil
				})
			},
		},
		{
			Name:        "degree sort",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					former := l.lessFunc
					defer func(former func(a, b int) bool) {
						l.lessFunc = former
					}(former)

					d, err := c.directedGraph()
					if err != nil {
						return err
					}
					l.lessFunc = func(a, b int) bool {
						aItem, _ := l.GetItem(a)
						aIn, _ := (*d).Incoming(aItem)
						aOut, _ := (*d).Outgoing(aItem)

						bItem, _ := l.GetItem(b)
						bIn, _ := (*d).Incoming(bItem)
						bOut, _ := (*d).Outgoing(bItem)

						return len(aIn)+len(aOut) < len(bIn)+len(bOut)
					}
					sort.Sort(l)
					return nil
				})
			},
		},
		{
			Name:        "delete Nodes",
			Description: "",
			run: func(c *command) error {
				nd, err := c.nodeDeleter()
				if err != nil {
					return err
				}
				return c.holderList(func(l *holderList) error {
					curIndex, err := l.GetCursorIndex()
					if err != nil {
						return err
					}
					var removed int
					for i := 0; i < l.Len(); i++ {
						l.Model.UpdateItem(i, func(item fmt.Stringer) (fmt.Stringer, error) {
							h, ok := item.(holder)
							if !ok {
								return h, fmt.Errorf("want %T, but got %T", h, item)
							}
							if !h.selected {
								return h, nil
							}

							err := (*nd).NodeDelete(h.content)
							if err != nil {
								return nil, err
							}
							if i < curIndex {
								curIndex--
							}
							i--
							removed++
							return nil, nil
						})
					}
					if removed == 0 {
						cur, err := l.GetCursorItem()
						if err != nil {
							return err
						}
						err = (*nd).NodeDelete(cur)
						if err != nil {
							return err
						}
						_, err = l.RemoveIndex(curIndex)
						return err
					}
					_, err = l.SetCursor(curIndex)

					return err
				})
			},
		},
		{
			Name:        "regex filter graph",
			Description: "",
			run: func(c *command) error {
				input, err := c.input("enter regex")
				if err != nil {
					return err
				}
				reg, err := regexp.Compile(input.String())
				if err != nil {
					return err
				}
				d, err := c.graphDirected()
				if err != nil {
					return err
				}
				c.returnGraph(&filter{origin: (*d), reg: reg})
				return nil
			},
		},
		{
			Name:        "filter main list",
			Description: "",
			run: func(c *command) error {
				input, err := c.input("enter filter string")
				if err != nil {
					return err
				}
				return c.holderList(func(l *holderList) error {
					all, err := l.GetAllItems()
					if err != nil {
						return err
					}
					filtered := make([]fmt.Stringer, 0, len(all))
					for _, item := range all {
						if strings.Contains(item.String(), input.String()) {
							filtered = append(filtered, item)
						}
					}
					return l.ResetItems(filtered...)
				})
			},
		},
		{
			Name:        "search main list",
			Description: "",
			run: func(c *command) error {
				i, err := c.input("enter search term")
				if err != nil {
					return err
				}
				input := i.String()
				return c.holderList(func(l *holderList) error {
					all, err := l.GetAllItems()
					if err != nil {
						return err
					}
					last, err := l.GetCursorIndex()
					if err != nil {
						return err
					}
					last++
					for i := last; i < l.Len(); i++ {
						if strings.Contains(all[i].String(), input) {
							_, err := l.SetCursor(i)
							return err
						}
					}
					for i := 0; i < last && i < l.Len(); i++ {
						if strings.Contains(all[i].String(), input) {
							_, err := l.SetCursor(i)
							return err
						}
					}
					return fmt.Errorf("no node found") // TODO explain that nodes can give different strings per call
				})
			},
		},
		{
			Name:        "navigator",
			Description: "",
			run: func(c *command) error {
				c.returnGraph(&navigator{})
				return nil
			},
		},
		{
			Name:        "create typed node",
			Description: "",
			run: func(c *command) error {
				t, err := c.nodeTypedCreator()
				if err != nil {
					return err
				}
				types, err := (*t).GetTypes()
				if err != nil {
					return err
				}

				input, err := c.input("new name")
				if err != nil {
					return err
				}

				n, err := c.choose(types...)
				if err != nil {
					return err
				}
				v, ok := n.(stringer)
				if !ok {
					return fmt.Errorf("expecting '%T' but got '%T'", v, n)
				}
				node, err := (*t).NodeTypedCreate(v, input)
				if err != nil {
					return err
				}
				c.newNodes(node)
				return nil
			},
		},
		{
			Name:        "execute",
			Description: "",
			run: func(c *command) error {
				e, err := c.executer()
				if err != nil {
					return err
				}
				i, err := c.node("to execute")
				if err != nil {
					return err
				}
				return (*e).Execute(i)
			},
		},
		{
			Name:        "open nodes", // ...fmt.Stringer -> walder.Graph
			Description: "",
			run: func(c *command) error {
				cur, err := c.nodeList("open graph")
				if err != nil {
					return err
				}
				c.pause("navigate to dimension in which to open the nodes")
				o, err := c.nodeOpener()
				if err != nil {
					return err
				}
				g, err := (*o).NodeOpen(cur...)
				if err != nil {
					return err
				}
				c.returnGraph(g)
				return nil
			},
		},
		{
			Name:        "remove others", // fmt.Stringer -> bool
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					lenght := l.Len()
					if len(l.GetSelected()) == 0 {
						cursor, err := l.GetCursorItem()
						if err != nil {
							return err
						}
						return l.ResetItems(cursor)
					}
					for i := 0; i < lenght; i++ {
						err := l.Model.UpdateItem(i, func(item fmt.Stringer) (fmt.Stringer, error) {
							v, ok := item.(holder)
							if !ok {
								return nil, fmt.Errorf("not a holder")
							}
							if !v.selected {
								i--             // closure
								return nil, nil // delete
							}
							v.selected = !v.selected
							return v, nil
						})
						if err != nil {
							return err
						}
					}
					return nil
				})
			},
		},
		{
			Name:        "focus incoming",
			Description: "",
			run: func(c *command) error {
				gh, err := c.graphHolder()
				if err != nil {
					return err
				}
				gh.focus = inAddr
				return nil
			},
		},
		{
			Name:        "focus outgoing",
			Description: "",
			run: func(c *command) error {
				gh, err := c.graphHolder()
				if err != nil {
					return err
				}
				gh.focus = outAddr
				return nil
			},
		},
		{
			Name:        "invert selected",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					updateFunc := func(i fmt.Stringer) (fmt.Stringer, error) {
						h, ok := i.(holder)
						if !ok {
							return i, fmt.Errorf("expected holder")
						}
						h.selected = !h.selected
						return h, nil
					}
					for c := 0; c < l.Len(); c++ {
						err := l.Model.UpdateItem(c, updateFunc)
						if err != nil {
							return err
						}
					}
					return nil
				})
			},
		},
		{
			Name:        "deselected all",
			Description: "",
			run: func(c *command) error {
				return c.holderList(func(l *holderList) error {
					updateFunc := func(i fmt.Stringer) (fmt.Stringer, error) {
						h, ok := i.(holder)
						if !ok {
							return i, fmt.Errorf("expected holder")
						}
						h.selected = false
						return h, nil
					}
					for c := 0; c < l.Len(); c++ {
						err := l.Model.UpdateItem(c, updateFunc)
						if err != nil {
							return err
						}
					}
					return nil
				})
			},
		},
		{
			Name:        "list errors",
			Description: "",
			run: func(c *command) error {
				var all []fmt.Stringer
				c.walder.peek().editList(errorAddr, func(l *holderList) error {
					var err error
					all, err = l.GetAllItems()
					return err
				})
				g, err := DotDim{}.New()
				if err != nil {
					return err
				}
				nc, ok := g.(walder.NodeCreater)
				if !ok {
					return fmt.Errorf("want %T, but got %T", nc, g)
				}
				for _, a := range all {
					_, err := nc.NodeCreate(a)
					if err != nil {
						return err
					}
				}
				c.walder.Push(nc)
				return nil
			}},
		{
			Name:        "invert graph",
			Description: "",
			run: func(c *command) error {
				g := c.walder.peek().graph
				toInvert, ok := g.(walder.GraphDirected)
				if !ok {
					c.walder.addError(fmt.Errorf("want %T, but got %T", toInvert, g))
				}
				return c.walder.Push(inverter{toInvert})
			}},
		{
			Name:        "start reachable transfer",
			Description: "",
			run: func(c *command) error {
				cur, err := c.node("start")

				from, err := c.graphDirected()
				if err != nil {
					return err
				}
				c.pause("get graphCreator")
				to, err := c.graphCreater()
				if err != nil {
					return err
				}
				return Reachable(cur, (*from), (*to))

			}},
		{
			Name:        "set dimension",
			Description: "",
			run: func(c *command) error {
				d, err := c.dimensionChanger()
				if err != nil {
					return err
				}
				all, err := (*d).DimensionGetAll()
				if err != nil {
					return err
				}
				n, err := c.choose(all...)
				return (*d).DimensionSet(n)
			},
		},
		{
			Name:        "push remover",
			Description: "",
			run: func(c *command) error {
				d, err := c.graphDirected()
				if err != nil {
					return err
				}

				startNodes, err := c.nodeList("filter defining nodes")
				if err != nil {
					return err
				}

				c.returnGraph(&remover{
					origin:     *d,
					dim:        dimBool,
					startNodes: startNodes,
				})
				return nil
			},
		},
		{
			Name:        "transfer",
			Description: "",
			run: func(c *command) error {
				oldGraph, err := c.graphDirected()
				if err != nil {
					return err
				}
				nodes, err := c.nodeList("to transfer")
				if err != nil {
					return err
				}
				gw, err := c.graphCreater()
				if err != nil {
					return err
				}
				lookup := make(map[string]fmt.Stringer)
				for _, oldNode := range nodes {
					newNode, err := (*gw).NodeCreate(oldNode)
					if err != nil {
						return err
					}
					if oldNode == nil {
						return fmt.Errorf("recieved nil value")
					}
					lookup[oldNode.String()] = newNode
				}
				for _, oldNode := range nodes {
					newNode, ok := lookup[oldNode.String()] // TODO is this possible?
					if !ok {
						return fmt.Errorf("failed hash lookup of node: '%#v'", oldNode)
					}
					out, err := (*oldGraph).Outgoing(oldNode)
					if err != nil {
						return err
					}
					for _, oldOut := range out {
						newOut, ok := lookup[oldOut.String()]
						if !ok {
							var err error
							newOut, err = (*gw).NodeCreate(oldOut)
							if err != nil {
								return err
							}
							lookup[oldOut.String()] = newOut
						}
						err = (*gw).EdgeCreate(newNode, newOut)
						if err != nil {
							return err
						}
					}
				}
				return nil
			},
		},
	}
)
