package pkg

import "io"

type tagRaw struct {
	name string
	data []byte
}

type Tag struct {
	id string
}

func (t *Tag) Hash(wkdir string) ([]byte, error) {
	return nil, nil
}

func (t *Tag) Type() string {
	return "tag"
}

func parseTag(r io.Reader, g *Got) (*Tag, error) {
	return nil, nil
}
