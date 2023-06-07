package lib

import (
	"fmt"

	boxer "github.com/treilik/bubbleboxer"
	"github.com/treilik/walder"
)

func newSparseModus(g walder.Graph) graphHolder {
	b := boxer.Boxer{}
	h := graphHolder{
		focus: mainAddr,
		graph: g,
		boxer: &b,
		updateFunc: func(g *graphHolder, w walder.Graph) {
			g.editList(inAddr, func(l *holderList) error { return l.ResetItems() })
			g.editList(outAddr, func(l *holderList) error { return l.ResetItems() })

			d, ok := w.(walder.GraphDirected)
			if !ok {
				return
			}

			current, err := g.getCursorItem()
			if err != nil {
				return
			}

			// limit path lenght
			limit := 50

			var o int
			outgoing := func(node fmt.Stringer) ([]fmt.Stringer, error) {
				o++
				if o > limit {
					return nil, nil
				}
				return d.Outgoing(node)
			}
			lastOutPos := func(current fmt.Stringer, from []fmt.Stringer) (int, error) {
				index, _ := g.lastOutPosition[current.String()]
				if index > len(from) {
					return 0, fmt.Errorf("last index is not possible")
				}
				return index, nil
			}
			beforeCurrent, err := choosePath(
				outgoing,
				lastOutPos,
				current,
			)
			if err != nil {
				return
			}

			var i int
			incoming := func(node fmt.Stringer) ([]fmt.Stringer, error) {
				i++
				if i > limit {
					return nil, nil
				}
				return d.Incoming(node)
			}
			lastInPos := func(current fmt.Stringer, from []fmt.Stringer) (int, error) {
				index, _ := g.lastInPosition[current.String()]
				if index > len(from) {
					return 0, fmt.Errorf("last index is not possible")
				}
				return index, nil
			}
			afterCurrent, err := choosePath(
				incoming,
				lastInPos,
				current,
			)
			if err != nil {
				return
			}

			lenght := len(beforeCurrent)
			reverse := make([]fmt.Stringer, 0, lenght)
			for c := lenght; c > 0; {
				c--
				reverse = append(reverse, beforeCurrent[c])
			}

			all := append(append(reverse, current), afterCurrent...)

			g.editList(mainAddr, func(l *holderList) error {
				err := l.ResetItems(all...)
				l.SetCursor(len(beforeCurrent))
				return err
			})

			mainListStrings := make(map[string]struct{})
			for _, a := range all {
				mainListStrings[a.String()] = struct{}{}
			}

			out, err := d.Outgoing(current)
			if err == nil && len(out) == 2 {
				func() {
					var o int
					outgoing := func(node fmt.Stringer) ([]fmt.Stringer, error) {
						o++
						if o > limit {
							return nil, nil
						}
						if _, ok := mainListStrings[node.String()]; ok {
							return nil, nil
						}
						return d.Outgoing(node)
					}
					lastOutPos := func(current fmt.Stringer, from []fmt.Stringer) (int, error) {
						index, _ := g.lastOutPosition[current.String()]
						if index > len(from) {
							return 0, fmt.Errorf("last index is not possible")
						}
						return index, nil
					}

					index, err := lastOutPos(current, out)
					if err != nil {
						return
					}
					other := out[(index+1)%2]
					outPath, err := choosePath(
						outgoing,
						lastOutPos,
						other,
					)
					if err != nil {
						return
					}
					outPath = append([]fmt.Stringer{other}, outPath...)
					g.editList(outAddr, func(l *holderList) error {
						return l.ResetItems(outPath...)
					})
				}()
			}

			in, err := d.Incoming(current)
			if err == nil && len(in) == 2 {
				func() {
					var i int
					incoming := func(node fmt.Stringer) ([]fmt.Stringer, error) {
						i++
						if i > limit {
							return nil, nil
						}
						if _, ok := mainListStrings[node.String()]; ok {
							return nil, nil
						}
						return d.Incoming(node)
					}
					lastInPos := func(current fmt.Stringer, from []fmt.Stringer) (int, error) {
						index, _ := g.lastInPosition[current.String()]
						if index > len(from) {
							return 0, fmt.Errorf("last index is not possible")
						}
						return index, nil
					}

					index, err := lastInPos(current, in)
					if err != nil {
						return
					}
					other := in[(index+1)%2]
					inPath, err := choosePath(
						incoming,
						lastInPos,
						other,
					)
					if err != nil {
						return
					}
					inPath = append([]fmt.Stringer{other}, inPath...)
					g.editList(inAddr, func(l *holderList) error {
						return l.ResetItems(inPath...)
					})
				}()
			}
		},
	}

	h.boxer.LayoutTree = boxer.Node{
		Children: []boxer.Node{
			{
				SizeFunc:        func(node boxer.Node, h int) []int { return []int{1, h - 1} },
				VerticalStacked: true,
				Children: []boxer.Node{
					stripErr(b.CreateLeaf(inModusAddr, stringer("in list"))),
					stripErr(b.CreateLeaf(inAddr, newHolderList(&h))),
				},
			},
			{
				SizeFunc:        func(node boxer.Node, h int) []int { return []int{1, h - 1} },
				VerticalStacked: true,
				Children: []boxer.Node{
					stripErr(b.CreateLeaf(mainModusAddr, stringer("main list"))),
					stripErr(b.CreateLeaf(mainAddr, newHolderList(&h))),
				},
			},
			{
				SizeFunc:        func(node boxer.Node, h int) []int { return []int{1, h - 1} },
				VerticalStacked: true,
				Children: []boxer.Node{
					stripErr(b.CreateLeaf(outModusAddr, stringer("out list"))),
					stripErr(b.CreateLeaf(outAddr, newHolderList(&h))),
				},
			},
		},
	}
	return h
}

// choosePath walk through the list of all the first returned nodes of the gen Function untill an error or the limit is reached
func choosePath(
	generate func(fmt.Stringer) ([]fmt.Stringer, error),
	choose func(current fmt.Stringer, from []fmt.Stringer) (int, error),
	start fmt.Stringer,
) ([]fmt.Stringer, error) {

	if generate == nil || choose == nil || start == nil {
		return nil, fmt.Errorf("recived nil value")
	}

	current := start
	var list []fmt.Stringer
	for {
		nodes, err := generate(current)
		if err != nil {
			return list, err
		}
		if len(nodes) == 0 {
			return list, nil
		}

		next, err := choose(current, nodes)
		if err != nil {
			return list, err
		}
		if next < 0 || next > len(nodes) {
			return list, fmt.Errorf("invalid choice")
		}
		list = append(list, nodes[next])
		current = nodes[next]
	}
}
