package spi

import (
	"io"
	"time"
)

type RowsEncoderContext struct {
	Sink         Sink
	Rownum       bool
	Heading      bool
	TimeLocation *time.Location
	TimeFormat   string
	Precision    int
}

type RowsEncoder interface {
	Open(colums Columns) error
	Close()
	AddRow(values []any) error
	Flush(heading bool)
	ContentType() string
}

type RowsDecoderContext struct {
	Reader       io.Reader
	TableName    string
	Columns      Columns
	TimeLocation *time.Location
	TimeFormat   string
}

type RowsDecoder interface {
	NextRow() ([]any, error)
}
