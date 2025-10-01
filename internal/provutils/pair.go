package provutils

import "fmt"

// Pair is a struct with two things.
type Pair[KA any, KB any] struct {
	A KA
	B KB
}

// NewPair creates a new Pair containing the provided values.
func NewPair[KA any, KB any](a KA, b KB) *Pair[KA, KB] {
	return &Pair[KA, KB]{A: a, B: b}
}

// SetA updates this Pair to have the provided value in A.
func (p *Pair[KA, KB]) SetA(a KA) {
	p.A = a
}

// GetA returns the A value in this Pair.
func (p *Pair[KA, KB]) GetA() KA {
	return p.A
}

// SetB updates this Pair to have the provided value in B.
func (p *Pair[KA, KB]) SetB(b KB) {
	p.B = b
}

// GetB returns the B value in this Pair.
func (p *Pair[KA, KB]) GetB() KB {
	return p.B
}

// Values returns both of the values in this Pair: A, B.
func (p *Pair[KA, KB]) Values() (KA, KB) {
	return p.A, p.B
}

// String returns a string representation of this Pair.
func (p *Pair[KA, KB]) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("<%v:%v>", p.A, p.B)
}
