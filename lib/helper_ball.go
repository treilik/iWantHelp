package lib

import (
	"fmt"

	"github.com/treilik/walder"
)

type ball struct {
	origin fmt.Stringer
	super  walder.GraphDirected
	size   int

	seen map[string]ballHolder
}

type ballHolder struct {
	distance int
	fmt.Stringer
}

var _ walder.GraphDirected = ball{}

func (b ball) String() string {
	if b.super == nil {
		return "empty ball"
	}
	return fmt.Sprintf("ball of '%s', with size '%d'", b.super.String(), b.size)
}
func (b ball) HomeNodes() ([]fmt.Stringer, error) {
	if b.origin == nil {
		return nil, fmt.Errorf("no origin node set")
	}
	return []fmt.Stringer{ballHolder{0, b.origin}}, nil
}
func (b ball) Incoming(node fmt.Stringer) ([]fmt.Stringer, error) {
	if b.origin == nil {
		return nil, fmt.Errorf("no origin node set")
	}
	bh, ok := node.(ballHolder)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", bh, node)
	}
	in, err := b.super.Incoming(bh.Stringer)
	if err != nil {
		return in, err
	}
	return b.wrap(bh, false, in...)
}
func (b ball) Outgoing(node fmt.Stringer) ([]fmt.Stringer, error) {
	if b.origin == nil {
		return nil, fmt.Errorf("no origin node set")
	}
	bh, ok := node.(ballHolder)
	if !ok {
		return nil, fmt.Errorf("want %T, but got %T", bh, node)
	}
	out, err := b.super.Outgoing(bh.Stringer)
	if err != nil {
		return out, err
	}
	return b.wrap(bh, true, out...)
}

func (b ball) wrap(from ballHolder, out bool, nodes ...fmt.Stringer) ([]fmt.Stringer, error) {
	if b.seen == nil {
		b.seen = make(map[string]ballHolder)
		b.seen[b.origin.String()] = ballHolder{0, b.origin}
	}
	holders := make([]fmt.Stringer, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			return nil, fmt.Errorf("recieved nil value")
		}
		str := n.String()
		if v, ok := b.seen[str]; ok && v.distance <= b.size {
			direction := 1
			if !out {
				direction = -1
			}

			holders = append(holders, ballHolder{v.distance + direction, v})
			continue
		}
		distance := from.distance + 1
		if distance > b.size {
			continue
		}
		New := ballHolder{distance, n}
		b.seen[str] = New
		holders = append(holders, New)
	}
	return holders, nil
}
