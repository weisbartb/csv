package csv

// UnmarshalCSV provides an interface to decode a non-scalar csv field.
// The string value will always be provided with any CSV related escaping already removed.
// If this returns an error, it will stop parsing for the record.
type UnmarshalCSV interface {
	UnmarshalCSV(data string) error
}

// MarshalCSV provides an interface to encode non-scalar values to a CSV.
type MarshalCSV interface {
	MarshalCSV() (string, error)
}

// Stringer provides an interface for a custom conversion to a string.
type Stringer interface {
	String() string
}

// Zeroer provides an interface to check if an object is in its zero state.
type Zeroer interface {
	IsZero() bool
}
