package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type finiteDeleter interface {
	walder.NodeAller
	walder.GraphOutgoing
	walder.EdgeDeleter
}

func reduction(graph finiteDeleter) error {
	all, err := graph.NodeAll()
	if err != nil {
		return err
	}

	for _, a := range all {
		found := make(map[string]fmt.Stringer)
		visit := func(path []fmt.Stringer) (abort bool, err error) {
			switch lenght := len(path); lenght {
			case 0, 1:
				return false, nil
			case 2:
				node := path[lenght-1]
				found[node.String()] = node
				return false, nil
			case 3:
				to, ok := found[path[lenght-1].String()]
				if ok {
					graph.EdgeDelete(a, to)
				}
			}
			return true, nil
		}
		BFS(graph, visit, a)
	}
	return nil
}
