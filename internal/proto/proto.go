package proto

import "fmt"



type ProtoErr struct {
	ErrString string
	inner error	
}

var (
	NetworkErr = &ProtoErr{ErrString: "Network Error: "}
)

func (p *ProtoErr) Error() string {
	return fmt.Sprintf("Protocol Error: %v\nInner:%s", p.ErrString,p.inner)
}

func (p *ProtoErr) Unwrap() error {
	return p.inner
}
