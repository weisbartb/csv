package csv

type UnmarshalCSV interface {
	UnmarshalCSV(data string) error
}
type MarshalCSV interface {
	MarshalCSV() (string, error)
}

type Stringer interface {
	String() string
}
