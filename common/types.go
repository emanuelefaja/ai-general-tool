package common

import "time"

// DataType represents the detected type of a column
type DataType string

const (
	TypeString  DataType = "string"
	TypeNumber  DataType = "number"
	TypeDate    DataType = "date"
	TypeBoolean DataType = "boolean"
	TypeMixed   DataType = "mixed"
	TypeEmpty   DataType = "empty"
)

// ColumnInfo contains metadata about a column
type ColumnInfo struct {
	Index        int
	Name         string
	DataType     DataType
	UniqueCount  int
	NullCount    int
	TotalCount   int
	SampleValues []string // First few unique values
}

// DataPreview represents the data structure for displaying file contents
type DataPreview struct {
	FileName     string
	FileType     string
	SheetInfo    string // For Excel files
	TotalRows    int
	TotalColumns int
	RowsDisplayed int
	SampleType   string // "first", "random"
	Columns      []ColumnInfo
	Headers      []string
	Rows         [][]string
}

// ParsedDate represents a parsed date value
type ParsedDate struct {
	Value time.Time
	Valid bool
}