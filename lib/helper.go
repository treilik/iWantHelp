package lib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type stringerGraph struct {
	list []fmt.Stringer
}

func (s stringerGraph) String() string {
	return "stringerGraph"
}
func (s stringerGraph) HomeNodes() ([]fmt.Stringer, error) {
	return s.list, nil
}

type stringer string

func (s stringer) String() string {
	return string(s)
}

func (s stringer) Init() tea.Cmd                         { return nil }
func (s stringer) Update(_ tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s stringer) View() string {
	return s.String()
}

type stringSorter []fmt.Stringer

func (s *stringSorter) Len() int           { return len(*s) }
func (s *stringSorter) Less(a, b int) bool { return (*s)[a].String() < (*s)[b].String() }
func (s *stringSorter) Swap(a, b int)      { (*s)[a], (*s)[b] = (*s)[b], (*s)[a] }

func sortStringer(toSort []fmt.Stringer) []fmt.Stringer {
	// TODO make pointer argument
	ts := stringSorter(toSort)
	sort.Sort(&ts)
	return ts
}

type buffer struct {
	bytes.Buffer
	path string
}

var _ io.WriteCloser = &buffer{}

func (b *buffer) Close() error {
	f, err := os.OpenFile(b.path, os.O_WRONLY, 644)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, b)
	if err != nil {
		return err
	}
	return f.Close()
}

func getNewWriter() func() (io.WriteCloser, error) {
	path := "/home/kili/.walder/" + strconv.FormatInt(time.Now().UnixMicro(), 10)
	os.Create(path)
	return func() (io.WriteCloser, error) {
		return &buffer{path: path}, nil
	}
}
