package lib

import (
	"fmt"

	"github.com/muesli/termenv"
	boxer "github.com/treilik/bubbleboxer"
	"github.com/treilik/walder"
)

const (
	inErr   = "inErr"
	mainErr = "mainErr"
	outErr  = "outErr"
)

func newTreeModus(g walder.Graph) graphHolder {

	b := boxer.Boxer{}
	h := graphHolder{
		focus: mainAddr,
		graph: g,
		boxer: &b,
		updateFunc: func(g *graphHolder, w walder.Graph) {
			defer g.setFocus()
			in(g)
			out(g)
		}}

	inHolder := newHolderList(&h)
	inHolder.CurrentStyle = termenv.Style{}

	outHolder := newHolderList(&h)
	outHolder.CurrentStyle = termenv.Style{}

	h.boxer.LayoutTree = boxer.Node{
		// root node

		Children: []boxer.Node{
			{
				VerticalStacked: true,
				Children: []boxer.Node{
					stripErr(b.CreateLeaf(inModusAddr, stringer("parent siblings"))),
					stripErr(b.CreateLeaf(inAddr, inHolder)),
					stripErr(b.CreateLeaf(inErr, inHolder)),
				},
				SizeFunc: func(node boxer.Node, h int) []int { return []int{1, h - 3, 2} },
			},
			{
				VerticalStacked: true,
				Children: []boxer.Node{
					stripErr(b.CreateLeaf(mainModusAddr, stringer("current siblings"))),
					stripErr(b.CreateLeaf(mainAddr, newHolderList(&h))),
					stripErr(b.CreateLeaf(mainErr, newHolderList(&h))),
				},
				SizeFunc: func(node boxer.Node, h int) []int { return []int{1, h - 3, 2} },
			},
			{
				VerticalStacked: true,
				Children: []boxer.Node{
					stripErr(b.CreateLeaf(outModusAddr, stringer("children"))),
					stripErr(b.CreateLeaf(outAddr, outHolder)),
					stripErr(b.CreateLeaf(outErr, outHolder)),
				},
				SizeFunc: func(node boxer.Node, h int) []int { return []int{1, h - 3, 2} },
			},
		}}

	return h
}
func in(g *graphHolder) {
	// in case a error occures reset to empty list
	g.resetIncoming()

	dir, ok := g.graph.(walder.GraphDirected)
	if !ok {
		g.resetIncoming(walder.NewError(fmt.Errorf("want %s, but got %T", graphDirectedString, g.graph)))
		return
	}
	cur, err := g.getCursorItem()
	if err != nil {
		g.resetIncoming(walder.NewError(err))
		return
	}
	parentList, err := dir.Incoming(cur)
	if err != nil {
		g.resetIncoming(walder.NewError(err))
		return
	}
	parLen := len(parentList)

	// not a tree
	if parLen > 1 {
		g.resetIncoming(walder.NewError(fmt.Errorf("not a tree")))
		return
	}

	// current is root node
	if parLen == 0 {
		return
	}

	parent := parentList[0]

	// to get the siblings of the parent we need to ask about the grandparent
	grandParentList, err := dir.Incoming(parent)
	if err != nil {
		g.resetIncoming(walder.NewError(err))
		return
	}
	grandParLen := len(grandParentList)
	if grandParLen > 1 {
		g.resetIncoming(walder.NewError(fmt.Errorf("not a tree")))
		return
	}
	// if we cant find the grandparent we dont set any siblings
	if grandParLen == 0 {
		g.resetIncoming(parentList...)
		return
	}

	// now we can ask the grandparent for all parentsiblings
	parentSiblings, err := dir.Outgoing(grandParentList[0])
	if err != nil {
		g.resetIncoming(walder.NewError(err))
		return
	}
	// set parentsiblings allready to be able to set the parent index
	g.resetIncoming(parentSiblings...)

	// check if tree modus is possible
	collisionFinder := make(map[string]struct{})

	parentString := parent.String()
	for i, c := range parentSiblings {
		childString := c.String()
		_, ok := collisionFinder[childString]
		if ok {
			// TODO snap out of tree modus
			g.resetIncoming(walder.NewError(fmt.Errorf("cant use string of nodes to uniqly identify")))
			return
		}
		collisionFinder[childString] = struct{}{}
		if childString == parentString {
			err := g.editList(inAddr, func(l *holderList) error {
				_, err := l.SetCursor(i)
				return err
			})
			if err != nil {
				g.resetIncoming(walder.NewError(err))
			}
		}
	}
	return
}

func current(g *graphHolder) {
	cur, err := g.getCursorItem()

	// in case a error occures reset to empty list
	g.resetMain()
	if err != nil {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(err))
		})
		return
	}
	dir, ok := g.graph.(walder.GraphDirected)
	if !ok {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(fmt.Errorf("want %s, but got %T", graphDirectedString, g.graph)))
		})
		return
	}
	parent, err := dir.Incoming(cur)
	if err != nil {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(err))
		})
		return
	}

	if len(parent) > 1 {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(fmt.Errorf("not a tree")))
		})
		return
	}

	// root node
	if len(parent) == 0 {
		g.resetMain(cur)
		return
	}

	children, err := dir.Outgoing(parent[0])
	if err != nil {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(err))
		})
		return
	}

	curString := cur.String()
	for i, c := range children {
		if c.String() == curString {
			g.editList(mainAddr, func(l *holderList) error {
				l.ResetItems(children...)
				_, err := l.SetCursor(i)
				return err
			})
			return
		}
	}
	_ = g.editList(mainErr, func(l *holderList) error {
		return l.AddItems(walder.NewError(fmt.Errorf("not a tree: current is not a child of its parent")))
	})

	return
}

func out(g *graphHolder) {
	// in case a error occures reset to empty list
	g.resetOutgoing()

	cur, err := g.getCursorItem()
	if err != nil {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(err))
		})
		return
	}

	dir, ok := g.graph.(walder.GraphDirected)
	if !ok {
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(fmt.Errorf("want %s, but got %T", graphDirectedString, g.graph)))
		})
		return
	}

	out, err := dir.Outgoing(cur)
	if err != nil {
		g.resetOutgoing(nil)
		_ = g.editList(mainErr, func(l *holderList) error {
			return l.AddItems(walder.NewError(err))
		})
		return
	}

	g.resetOutgoing(out...)
	return
}

// resetIncoming resets all items in the incoming list in the layout tree
// but only if the stringerList is not nil, otherwise it stays the same.
func (b *graphHolder) resetIncoming(stringerList ...fmt.Stringer) {
	if stringerList == nil {
		stringerList = []fmt.Stringer{}
	}
	err := b.editList(inAddr, func(l *holderList) error {
		return l.ResetItems(stringerList...)
	})
	if err == nil {
		return
	}
	_ = b.editList(inErr, func(l *holderList) error {
		return l.ResetItems(walder.NewError(err))
	})
}

// resetMain resets all items in the main list in the layout tree
// but only if the stringerList is not nil, otherwise it stays the same.
func (b *graphHolder) resetMain(stringerList ...fmt.Stringer) {
	if stringerList == nil {
		stringerList = []fmt.Stringer{}
	}
	err := b.editList(mainAddr, func(l *holderList) error {
		return l.ResetItems(stringerList...)
	})
	if err == nil {
		return
	}
	_ = b.editList(mainErr, func(l *holderList) error {
		return l.AddItems(walder.NewError(err))
	})
}

// resetOutgoing resets all items in the outcoming list in the layout tree
// but only if the stringerList is not nil, otherwise it stays the same.
func (b *graphHolder) resetOutgoing(stringerList ...fmt.Stringer) {
	if stringerList == nil {
		stringerList = []fmt.Stringer{}
	}
	err := b.editList(outAddr, func(l *holderList) error {
		return l.ResetItems(stringerList...)
	})
	if err == nil {
		return
	}
	_ = b.editList(outErr, func(l *holderList) error {
		return l.AddItems(walder.NewError(err))
	})
}

func (b *graphHolder) moveIncoming() error {
	var empty bool
	_ = b.editList(inAddr, func(l *holderList) error {
		if l.Len() == 0 {
			empty = true
		}
		return nil
	})
	if empty {
		return nil
	}

	mainModel := b.boxer.ModelMap[mainAddr]
	inModel := b.boxer.ModelMap[inAddr]
	b.boxer.ModelMap[mainAddr] = inModel
	b.boxer.ModelMap[outAddr] = mainModel
	return nil
}

func (b *graphHolder) moveOutgoing() error {
	var empty bool
	_ = b.editList(outAddr, func(l *holderList) error {
		if l.Len() == 0 {
			empty = true
		}
		return nil
	})
	if empty {
		return nil
	}

	mainModel := b.boxer.ModelMap[mainAddr]
	outModel := b.boxer.ModelMap[outAddr]
	b.boxer.ModelMap[mainAddr] = outModel
	b.boxer.ModelMap[inAddr] = mainModel
	return nil
}

func (b *graphHolder) getCursorItem() (fmt.Stringer, error) {
	mainModel := b.boxer.ModelMap[mainAddr]
	mainLister := mainModel.(holderList)
	return mainLister.GetCursorItem()
}

func (b *graphHolder) getSelected() []fmt.Stringer {
	mainModel := b.boxer.ModelMap[mainAddr]
	mainLister := mainModel.(holderList)
	return mainLister.GetSelected()
}

func (b *graphHolder) getCurrent() ([]fmt.Stringer, error) {
	selected := b.getSelected()
	if len(selected) != 0 {
		return selected, nil
	}
	cursor, err := b.getCursorItem()
	return []fmt.Stringer{cursor}, err
}

func (b *graphHolder) editList(addr string, editFunc func(*holderList) error) error {
	// snap out of edge focus if main list is changed
	if b.focus != mainAddr && addr == mainAddr {
		b.focus = mainAddr
	}
	model, ok := b.boxer.ModelMap[addr]
	if !ok {
		return fmt.Errorf("address '%s' is not in the LayoutTree", addr)
	}
	listModel, ok := model.(holderList)
	if !ok {
		return fmt.Errorf("address '%s' is not a '%T'", addr, holderList{})
	}
	err := editFunc(&listModel)
	if err != nil {
		return err
	}
	b.boxer.ModelMap[addr] = listModel
	return nil
}
