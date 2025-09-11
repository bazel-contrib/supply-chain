package label

/*
 * Label implementation
 */
var _ Label = (*simple)(nil)

// simple is a simple implementation of `Label` using a simple string internally.
type simple struct {
	Label string
}

func (s *simple) labelPrivate() {}

func (s *simple) String() string {
	return s.Label
}
