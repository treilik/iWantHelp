package lib

import (
	"fmt"
	"regexp"

	"github.com/treilik/walder"
)

type Dim struct{}

func (d Dim) String() string {
	return "filter"
}

var _ walder.Wraper = Dim{}

func (d Dim) Wrap(origin walder.Graph) (walder.Graph, error) {
	dir, ok := origin.(walder.GraphDirected)
	if !ok {
		return nil, fmt.Errorf("the given graph dos not satisfy the %s interface", graphDirectedString)
	}
	return &filter{origin: dir}, nil
}

type filter struct {
	origin walder.GraphDirected
	reg    *regexp.Regexp
}

var _ walder.GraphDirected = &filter{}

func (f *filter) String() string {
	if f.origin == nil {
		return "nothing to filter" // TODO fix error through string
	}
	return fmt.Sprintf("Filter of Graph '%s' with '%s'", f.origin.String(), f.reg.String())
}

func (f *filter) filter(toFilter []fmt.Stringer) ([]fmt.Stringer, error) {
	filtered := make([]fmt.Stringer, 0, len(toFilter))
	for _, i := range toFilter {
		if i == nil {
			continue // TODO handel nil values as errors? configurable
		}
		if f.reg.MatchString(i.String()) {
			filtered = append(filtered, i)
		}
	}
	return filtered, nil
}

func (f *filter) HomeNodes() ([]fmt.Stringer, error) {
	if f.origin == nil {
		return nil, fmt.Errorf("nothing to filter")
	}
	home, err := f.origin.HomeNodes()
	if err != nil {
		return nil, err
	}
	return f.filter(home)
}

func (f *filter) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	if f.origin == nil {
		return nil, fmt.Errorf("nothing to filter")
	}

	in, err := f.origin.Incoming(node)
	if err != nil {
		return in, err
	}
	return f.filter(in)
}

func (f *filter) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	if f.origin == nil {
		return nil, fmt.Errorf("nothing to filter")
	}
	in, err := f.origin.Outgoing(node)
	if err != nil {
		return in, err
	}
	return f.filter(in)
}

var _ walder.Meta = &filter{}

func (f *filter) Get() (walder.Graph, error) {
	return f.origin, nil
}
func (f *filter) Set(g walder.Graph) error {
	d, ok := g.(walder.GraphDirected)
	if !ok {
		return fmt.Errorf("cant use %T as %s", g, graphDirectedString)
	}
	f.origin = d
	return nil
}
