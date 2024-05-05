# Structured CSV Tools
This library provides a set of methods for encoding and decoding CSVs from/into objects.

## Example
```go
fh, err := os.Open("testdata/simple.csv")
if err != nil {
  return err
}
reader, err := csv.NewStructuredCSVReader[myDataStruct](fh)
for {
  record, err = reader.Next()
  if err != nil {
      if errors.Is(err, io.EOF) {
          break
      }
      return err
  }
}
```

## Encoder/Decoder Options
The encoder and decoder support a few inline options once the reader/writer is created.
These are exported fields that can be altered to change the behavior of the reader/writer being altered.
More options will be added over time

### Encoder
- `StrictMode` turns on strict parsing rules, unknown fields will be reported as an error on the first row read.

## Tag Format

This library uses struct tags to pull data about the field for decoding.
Fields without tags **will not** be parsed.
All fields must be exported to used by this library, non-exported fields are automatically skipped.

### Tag Format
Tags are formatted as such: `"csv:<fieldName>,[required,][omitempty,]"`

- `<fieldName>` represents the name of the CSV field, this should be in the header row.
- required is a parameter that when present causes the field to error if its null when decoding the value
- omitempty is a parameter that when present causes empty values to encode to null.
    - `Zeroer` is a supported interface to detect a zero value, otherwise the default reflection zero detection is used.


## Value Encoding/Decoding

This library provides native support for scalar values.
Non-scalar values can be decoded by implementing supported interfaces (covered below).
Interfaces are checked in the order they are specified in

### Non-scalar Decoding

1. `UnmarshalCSV`
2. `encoding.TextUnmarshaler`


### Non-scalar Encoding

1. `MarshalCSV`
2. `encoding.Text<arshaler`
3. `Stringer`

## Caveats

### Omit Empty

#### Non-nullables

The zero value is used for any non nullable value when a null value is provided.

#### Strings

Empty strings that are not encapsulated in `""` are considered null, by default most CSV writers will not do this.
If your application depends on empty string values,
you should prepare to handle nulls and appropriately handle zero values.