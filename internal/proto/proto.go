package proto

import (
	"fmt"
)

type ProtoErr struct {
	Context string
	Inner   error
}

var (
	GenericNetErr = &ProtoErr{Context: "Network Error"}
)

func (p *ProtoErr) Error() string {
	if p.Inner != nil {
		return fmt.Sprintf("Protocol Error: %v\nInner:%s", p.Context, p.Inner)
	}
	return fmt.Sprintf("Protocol Error: %s\n", p.Context)
}

func (p *ProtoErr) Unwrap() error {
	return p.Inner
}
