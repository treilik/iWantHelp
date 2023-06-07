package lib

import (
	md5 "crypto/md5"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	lister "github.com/treilik/bubblelister"
	"github.com/treilik/walder"
)

var (
	style = termenv.Style{}.Foreground(termenv.ANSIRed)
)

type holder struct {
	g        *graphHolder
	content  fmt.Stringer
	selected bool
	labels   [][2]string
}

func newHolder(g *graphHolder, c fmt.Stringer) holder {
	return holder{
		g:       g,
		content: c,
	}
}

func (h holder) String() string {
	content := h.content.String()
	if h.labels != nil {
		tmp := make([]string, 0, len(h.labels)+1)
		tmp = append(tmp, content)
		for _, kv := range h.labels {
			tmp = append(tmp, fmt.Sprintf("  %s: %s", kv[0], kv[1]))
		}
		content = strings.Join(tmp, "\n")
	}
	// color according to type
	if h.g != nil {
		t, ok := h.g.graph.(walder.Typer)
		if ok {
			T, err := t.GetType(h.content)
			if err == nil {
				hash := fmt.Sprintf("#%6x", md5.Sum([]byte(T)))
				p := termenv.ANSI
				c := p.Color(hash)
				content = termenv.Style{}.Foreground(c).Styled(content)

			}
		}
	}
	// overwrite type // TODO change
	if h.selected {
		return fmt.Sprintf("  %s", style.Styled(content))
	}
	return content
}

type holderList struct {
	g *graphHolder
	lister.Model
	lessFunc func(a, b int) bool
}

func (h *holderList) Less(a, b int) bool {
	if h.lessFunc != nil {
		return h.lessFunc(a, b)
	}
	aItem, _ := h.Model.GetItem(a)
	bItem, _ := h.Model.GetItem(b)
	return aItem.String() < bItem.String()
}

func newHolderList(g *graphHolder) holderList {
	return holderList{g, lister.NewModel(), nil}
}

func (l holderList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := l.Model.Update(msg)
	newList := newModel.(lister.Model)
	l.Model = newList
	return l, cmd
}

func (l *holderList) GetAllItems() ([]fmt.Stringer, error) {
	allItems := l.Model.GetAllItems()
	all := make([]fmt.Stringer, 0, len(allItems))
	for _, i := range allItems {
		h, ok := i.(holder)
		if !ok {
			return nil, fmt.Errorf("a wrong type was in the list: %T", i)
		}
		all = append(all, h.content)
	}

	return all, nil
}

func (l *holderList) AddItems(items ...fmt.Stringer) error {
	holderItems := make([]fmt.Stringer, 0, len(items))
	for _, i := range items {
		if i == nil {
			return lister.NilValue(fmt.Errorf("wont add nil value to list"))
		}
		holderItems = append(holderItems, newHolder(l.g, i))
	}
	return l.Model.AddItems(holderItems...)
}

func (l *holderList) UpdateItem(index int, updateFunc func(fmt.Stringer) (fmt.Stringer, error)) error {
	closureFunc := func(item fmt.Stringer) (fmt.Stringer, error) {
		h, ok := item.(holder)
		if !ok {
			return h, fmt.Errorf("not a holder: %T", item)
		}
		newContent, err := updateFunc(h.content)
		if err != nil {
			return h, err
		}
		h.content = newContent
		return h, nil

	}
	return l.Model.UpdateItem(index, closureFunc)
}
func (l *holderList) ResetItems(newStringers ...fmt.Stringer) error {
	newHolders := make([]fmt.Stringer, 0, len(newStringers))
	for _, s := range newStringers {
		if s == nil {
			return fmt.Errorf("recieved nil value")
		}
		newHolders = append(newHolders, newHolder(l.g, s))
	}
	return l.Model.ResetItems(newHolders...)
}
func (l *holderList) GetCursorItem() (fmt.Stringer, error) {
	i, err := l.Model.GetCursorItem()
	if err != nil {
		return i, err
	}
	h, ok := i.(holder)
	if !ok {
		return h, fmt.Errorf("not a holder: %T", i)
	}
	return h.content, nil
}
func (l *holderList) GetItem(index int) (fmt.Stringer, error) {
	i, err := l.Model.GetItem(index)
	if err != nil {
		return i, err
	}
	h, ok := i.(holder)
	if !ok {
		return h, fmt.Errorf("not a holder: %T", i)
	}
	return h.content, nil
}
func (l *holderList) LessFunc(a, b fmt.Stringer) bool {
	h, ok := a.(holder)
	if ok {
		a = h.content
	}
	h, ok = b.(holder)
	if ok {
		b = h.content
	}
	return l.Model.LessFunc(a, b)
}
func (l *holderList) EqualsFunc(a, b fmt.Stringer) bool {
	h, ok := a.(holder)
	if ok {
		a = h.content
	}
	h, ok = b.(holder)
	if ok {
		b = h.content
	}
	return l.Model.EqualsFunc(a, b)
}

func (l *holderList) GetSelected() []fmt.Stringer {
	var selectedList []fmt.Stringer
	for _, i := range l.Model.GetAllItems() {
		h := i.(holder)
		if !h.selected {
			continue
		}
		selectedList = append(selectedList, h.content)
	}
	return selectedList
}

func (l *holderList) SetSelect(index int, value bool) error {
	return l.Model.UpdateItem(index, func(i fmt.Stringer) (fmt.Stringer, error) {
		h := i.(holder)
		h.selected = value
		return h, nil
	})
}
func (l *holderList) GetSelect(index int) (bool, error) {
	i, err := l.Model.GetItem(index)
	if err != nil {
		return false, err
	}
	h := i.(holder)
	return h.selected, nil
}
