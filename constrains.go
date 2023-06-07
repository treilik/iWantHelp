package walder

type Constrain string

const (
	connected = Constrain("connected")
	tree      = Constrain("tree")
	dag       = Constrain("dag")
	strict    = Constrain("strict")
)

var GraphConstrainAllowed = map[Constrain]string{
	connected: "",
	dag:       "",
	strict:    "",
	tree:      "",
}
