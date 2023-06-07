package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type Visitor func([]fmt.Stringer) (abort bool, err error)

func DFS(directed walder.GraphOutgoing, visitor Visitor, start fmt.Stringer) error {
	return dfs(directed, visitor, start)
}
func dfs(directed walder.GraphOutgoing, visitor Visitor, start ...fmt.Stringer) error {
	if directed == nil {
		return fmt.Errorf("recived nil")
	}
	if start == nil || len(start) == 0 {
		return fmt.Errorf("recived nil")
	}
	if visitor == nil {
		return fmt.Errorf("recived nil")
	}
	out, err := directed.Outgoing(start[len(start)-1])
	if err != nil {
		return err
	}
	for _, o := range out {
		path := make([]fmt.Stringer, len(start))
		copy(path, start)
		path = append(path, o)
		abort, err := visitor(path)
		if err != nil {
			return err
		}
		if abort {
			continue
		}
		dfs(directed, visitor, path...)
	}

	return nil
}
func BFS(directed walder.GraphOutgoing, visitor Visitor, start fmt.Stringer) error {
	return bfs(directed, visitor, []fmt.Stringer{start})
}
func bfs(directed walder.GraphOutgoing, visitor Visitor, paths ...[]fmt.Stringer) error {
	if directed == nil {
		return fmt.Errorf("recived nil")
	}
	if paths == nil {
		return fmt.Errorf("recived nil")
	}
	if visitor == nil {
		return fmt.Errorf("recived nil")
	}
	var next [][]fmt.Stringer
	for _, path := range paths {
		if len(path) == 0 {
			return fmt.Errorf("recieved empty path")
		}
		abort, err := visitor(path)
		if err != nil {
			return err
		}
		if abort {
			continue
		}

		out, err := directed.Outgoing(path[len(path)-1])
		if err != nil {
			return err
		}
		for _, o := range out {
			next = append(next, append(path, o))
		}
	}
	bfs(directed, visitor, next...)
	return nil
}
