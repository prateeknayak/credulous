package cio

import (
	"fmt"
	"io"
)

type ConsoleWriter struct {
	io.Writer
}

func NewConsoleWriter(writer io.Writer) *ConsoleWriter {
	return &ConsoleWriter{
		writer,
	}
}

func (d *ConsoleWriter) Display(message string) {
	fmt.Fprintf(d.Writer, message)
}
