package eval

import (
	"fmt"
	"os"
)

type Pipe struct {
	r, w *os.File
}

var _ Value = Pipe{}

func (Pipe) Kind() string {
	return "pipe"
}

func (p Pipe) Eq(rhs interface{}) bool {
	return p == rhs
}

func (p Pipe) Repr(int) string {
	return fmt.Sprintf("<pipe{%v %v}>", p.r.Fd(), p.w.Fd())
}
