package lib

import (
	"fmt"

	"github.com/muesli/termenv"
	boxer "github.com/treilik/bubbleboxer"
	"github.com/treilik/walder"
)

func newDirectedBoxer(g walder.Graph) graphHolder {
	b := boxer.Boxer{}
	h := graphHolder{
		graph: g,
		focus: mainAddr,
		boxer: &b,
		updateFunc: func(b *graphHolder, g walder.Graph) {
			// set focus heighlighting
			defer b.setFocus()

			var in, out []fmt.Stringer
			_ = b.editList(mainAddr, func(l *holderList) error {
				cur, err := l.GetCursorItem()
				if err != nil {
					_ = b.editList(mainErr, func(l *holderList) error { return l.AddItems(walder.NewError(err)) })
					return nil
				}
				// TODO move type check into graphHolder instantiation
				dir, ok := b.graph.(walder.GraphDirected)
				if !ok {
					err := fmt.Errorf("want %s, but got %T", graphDirectedString, b.graph)
					_ = b.editList(mainErr, func(l *holderList) error { return l.AddItems(walder.NewError(err)) })
					return nil
				}
				in, err = dir.Incoming(cur)
				if err != nil {
					_ = b.editList(inErr, func(l *holderList) error { return l.AddItems(walder.NewError(err)) })
					return nil
				}
				out, err = dir.Outgoing(cur)
				if err != nil {
					_ = b.editList(outErr, func(l *holderList) error { return l.AddItems(walder.NewError(err)) })
					return nil
				}
				return nil
			})
			b.resetIncoming(in...)
			b.resetOutgoing(out...)
			return
		},
	}

	inHolder := newHolderList(&h)
	inHolder.CurrentStyle = termenv.Style{}

	outHolder := newHolderList(&h)
	outHolder.CurrentStyle = termenv.Style{}

	h.boxer.LayoutTree = boxer.Node{
		// root node
		VerticalStacked: true,
		Children: []boxer.Node{
			{
				// list node
				Children: []boxer.Node{
					{
						VerticalStacked: true,
						Children: []boxer.Node{
							stripErr(b.CreateLeaf(inModusAddr, stringer("incoming"))),
							stripErr(b.CreateLeaf(inAddr, inHolder)),
							stripErr(b.CreateLeaf(inErr, newHolderList(&h))),
						},
						SizeFunc: func(n boxer.Node, h int) []int { return []int{1, h - 3, 2} },
					},
					{
						VerticalStacked: true,
						Children: []boxer.Node{
							stripErr(b.CreateLeaf(mainModusAddr, stringer("undefined"))),
							stripErr(b.CreateLeaf(mainAddr, newHolderList(&h))),
							stripErr(b.CreateLeaf(mainErr, newHolderList(&h))),
						},
						SizeFunc: func(n boxer.Node, h int) []int { return []int{1, h - 3, 2} },
					},
					{
						VerticalStacked: true,
						Children: []boxer.Node{
							stripErr(b.CreateLeaf(outModusAddr, stringer("outgoing"))),
							stripErr(b.CreateLeaf(outAddr, outHolder)),
							stripErr(b.CreateLeaf(outErr, newHolderList(&h))),
						},
						SizeFunc: func(n boxer.Node, h int) []int { return []int{1, h - 3, 2} },
					},
				},
			},
		},
	}
	return h

}
