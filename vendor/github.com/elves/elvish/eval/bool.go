package eval

// Bool represents truthness.
type Bool bool

func (Bool) Kind() string {
	return "bool"
}

func (b Bool) Eq(rhs interface{}) bool {
	return b == rhs
}

func (b Bool) Repr(int) string {
	if b {
		return "$true"
	}
	return "$false"
}

func (b Bool) Bool() bool {
	return bool(b)
}

// ToBool converts a Value to bool. When the Value type implements Bool(), it
// is used. Otherwise it is considered true.
func ToBool(v Value) bool {
	if b, ok := v.(Booler); ok {
		return b.Bool()
	}
	return true
}
