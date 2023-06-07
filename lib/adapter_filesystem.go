package lib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/treilik/walder"
)

// FSDim is a generator for the Filesystem Dimension
type FSDim struct{}

var _ walder.Dimensioner = FSDim{}

func (d FSDim) String() string {
	return "Filesystem"
}

// New returns a new representation graph of the Filesystem
func (d FSDim) New() (walder.Graph, error) {
	return &FS{}, nil
}

var (
	_ walder.GraphDirected = &FS{}
)

type filePath string

var _ walder.Pather = filePath("")

func (f filePath) Path() (string, error) {
	return string(f), nil
}

func (f filePath) String() string {
	return filepath.Base(string(f))
}

type FS struct{}

func (f *FS) String() string {
	return "filesystem"
}

func (f *FS) HomeNodes() ([]fmt.Stringer, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return []fmt.Stringer{filePath(path)}, nil
}

func (f *FS) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	path, ok := node.(filePath)
	if !ok {
		return nil, fmt.Errorf("got '%T' but want '%T'", node, path)
	}

	dir := filepath.Dir(string(path))
	if dir == string(path) {
		return []fmt.Stringer{}, nil
	}
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	return []fmt.Stringer{filePath(dir)}, nil
}
func (f *FS) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	path, ok := node.(filePath)
	if !ok {
		return nil, fmt.Errorf("got '%T' but want '%T'", node, path)
	}
	fileList, _ := os.ReadDir(string(path))
	stringerList := make([]fmt.Stringer, 0, len(fileList))
	for _, f := range fileList {
		stringerList = append(stringerList, filePath(filepath.Join(string(path), f.Name())))
	}
	return stringerList, nil
}

var _ walder.NodeReader = &FS{}

func (f *FS) NodeRead(node fmt.Stringer) (io.Reader, error) {
	path, ok := node.(filePath)
	if !ok {
		return nil, fmt.Errorf("got '%T' but want '%T'", node, path)
	}
	content, err := os.ReadFile(string(path))
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(content)
	return reader, nil
}

var _ walder.NodeWriter = &FS{}

func (f *FS) NodeUpdate(n fmt.Stringer) (fmt.Stringer, error) {
	return n, nil
}

func (f *FS) NodeWrite(node fmt.Stringer) (io.WriteCloser, error) {
	path, ok := node.(filePath)
	if !ok {
		return nil, fmt.Errorf("got '%T' but want '%T'", node, path)
	}
	fd, err := os.OpenFile(string(path), os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}
	return fd, nil
}

var _ walder.NodeNamer = &FS{}

func (f *FS) NodeName(node fmt.Stringer, name string) (fmt.Stringer, error) {
	path, ok := node.(filePath)
	if !ok {
		return nil, fmt.Errorf("got '%T' but want '%T'", node, path)
	}
	parent := filepath.Dir(string(path))
	New := filepath.Join(parent, name)
	err := os.Rename(string(path), New)
	return filePath(New), err
}

var _ walder.Typer = &FS{}

func (f *FS) GetType(node fmt.Stringer) (string, error) {
	path, ok := node.(filePath)
	if !ok {
		return "", fmt.Errorf("got '%T' but want '%T'", node, path)
	}

	s, err := os.Stat(string(path))
	if err != nil {
		return "", err
	}
	if s.IsDir() {
		return "directory", nil
	}
	return "file", nil
}

type cmdNode struct {
	name string
	cmd  func(w *Walder, file filePath) error
}

func (c cmdNode) String() string {
	return c.name
}

var _ walder.NodeFromCreater = &FS{}

func (f *FS) NodeFromCreate(input fmt.Stringer, from ...fmt.Stringer) (fmt.Stringer, error) {
	if input == nil {
		return nil, fmt.Errorf("recieved nil value")
	}
	if len(from) != 1 {
		return nil, fmt.Errorf("cant yet create multiple hardlinks")
	}
	path, ok := from[0].(filePath)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", path, from[0])
	}

	s, err := os.Stat(string(path))
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("%s is not a directory, thus no file can be created in it", path)
	}

	newPath := filepath.Join(string(path), input.String())
	_, err = os.Create(newPath)
	if err != nil {
		return nil, err
	}
	return filePath(newPath), nil
}

var _ walder.EdgeMover = &FS{}

func (fs *FS) EdgeMove(toMove, from, to fmt.Stringer) error {
	m, ok := toMove.(filePath)
	if !ok {
		return fmt.Errorf("want %T, but got %T", m, toMove)
	}
	f, ok := from.(filePath)
	if !ok {
		return fmt.Errorf("want %T, but got %T", f, from)
	}
	t, ok := to.(filePath)
	if !ok {
		return fmt.Errorf("want %T, but got %T", t, to)
	}
	if filepath.Dir(string(m)) != string(f) {
		return fmt.Errorf("%s no child from %s", toMove.String(), from.String())
	}
	New := filepath.Join(string(t), filepath.Base(string(m)))
	return os.Rename(string(m), New)
}
