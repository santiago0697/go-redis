package main

import "io"

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)

	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) EmptyWrite() error {
	err := w.Write(EmptyValue())
	if err != nil {
		return err
	}

	return nil
}
