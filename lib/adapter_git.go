package lib

import (
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/treilik/walder"
)

type GitDim struct{}

var _ walder.NodeOpener = GitDim{}
var _ walder.Dimensioner = GitDim{}

func (d GitDim) String() string {
	return "git"
}
func (d GitDim) New() (walder.Graph, error) {
	return nil, fmt.Errorf("not yet implemnted")
}
func (d GitDim) NodeOpen(nodes ...fmt.Stringer) (walder.Graph, error) {
	if len(nodes) != 1 {
		return nil, fmt.Errorf("need exactly one node, got %d", len(nodes))
	}
	node := nodes[0]
	if node == nil {
		return nil, fmt.Errorf("recieved nil value")
	}

	path := node.String()

	if p, ok := node.(walder.Pather); ok {
		var err error
		path, err = p.Path()
		if err != nil {
			return nil, err
		}
	}
	r, err := git.PlainOpen(path)
	return Repo{r}, err
}

type Repo struct {
	*git.Repository
}

var _ walder.Graph = Repo{}

func (r Repo) String() string {
	return "Git Repo"
}

type commit object.Commit

func (c commit) String() string {
	line, _, _ := strings.Cut(c.Message, "\n")
	return fmt.Sprintf("%s %s", c.Hash.String()[0:7], line)
}

func (r Repo) HomeNodes() ([]fmt.Stringer, error) {
	if r.Repository == nil {
		return nil, fmt.Errorf("no repository set")
	}
	h, err := r.Head()
	if err != nil {
		return nil, err
	}
	c, err := r.CommitObject(h.Hash())
	return []fmt.Stringer{commit(*c)}, err
}

var _ walder.GraphDirected = Repo{}

func (r Repo) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	n, ok := node.(commit)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", n, node)
	}

	var stringers []fmt.Stringer
	for _, h := range n.ParentHashes {
		c, err := r.CommitObject(h)
		if err != nil {
			return nil, err
		}
		stringers = append(stringers, commit(*c))
	}
	return stringers, nil
}
func (r Repo) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	n, ok := node.(commit)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", n, node)
	}

	i, err := r.CommitObjects()
	if err != nil {
		return nil, err
	}

	var stringers []fmt.Stringer
	c, err := i.Next()
	for err == nil {
		for _, p := range c.ParentHashes {
			if p != n.Hash {
				continue
			}
			stringers = append(stringers, commit(*c))
		}
		c, err = i.Next()
	}
	return stringers, nil
}
