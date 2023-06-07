package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

func BreakCycles(start fmt.Stringer, toBreak walder.GraphOutgoing, writeInto walder.GraphCreater) error {
	seen := make(map[string]fmt.Stringer)
	visit := func(path []fmt.Stringer) (bool, error) {
		lenght := len(path)
		last := path[lenght-1]
		if lenght == 1 {
			New, err := writeInto.NodeCreate(last)
			if err != nil {
				return false, err
			}
			seen[New.String()] = New
			return false, nil
		}

		if _, ok := seen[last.String()]; ok {
			abort := true
			return abort, nil
		}
		from, ok := seen[path[lenght-2].String()]
		if !ok {
			return false, fmt.Errorf("node allready found but not found again")
		}
		to, err := writeInto.NodeCreate(last)
		if err != nil {
			return false, err
		}
		seen[last.String()] = to
		return false, writeInto.EdgeCreate(from, to)
	}
	return BFS(toBreak, visit, start)
}
