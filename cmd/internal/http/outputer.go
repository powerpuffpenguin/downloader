package http

import (
	"bytes"
	"fmt"
	"io"
)

type _Writer struct {
	writer io.Writer
	n      int
}

func (w *_Writer) Write(b []byte) (n int, err error) {
	n, err = w.writer.Write(b)
	if n != 0 {
		b = b[:n]
		find := bytes.LastIndexByte(b, '\n')
		if find == -1 {
			w.n += n
		} else {
			w.n += n - find - 1
		}
	}
	return
}

type Outputer struct {
	Writer  io.Writer
	OutLine int
}

func (o *Outputer) outLine(a ...interface{}) {
	if o.OutLine != 0 {
		fmt.Fprintf(o.Writer, fmt.Sprintf("\r%%-%ds\r", o.OutLine), "")
		o.OutLine = 0
	}
}

func (o *Outputer) Print(a ...interface{}) (n int, err error) {
	w := &_Writer{
		writer: o.Writer,
	}
	n, err = fmt.Fprint(w, a...)
	o.OutLine = w.n
	return
}
func (o *Outputer) PrintLine(a ...interface{}) (n int, err error) {
	o.outLine()
	n, err = o.Print(a...)
	return
}
func (o *Outputer) Println(a ...interface{}) (n int, err error) {
	w := &_Writer{
		writer: o.Writer,
	}
	n, err = fmt.Fprintln(w, a...)
	o.OutLine = w.n
	return
}
func (o *Outputer) PrintlnLine(a ...interface{}) (n int, err error) {
	o.outLine()
	n, err = o.Println(a...)
	return
}
func (o *Outputer) Printf(format string, a ...interface{}) (n int, err error) {
	w := &_Writer{
		writer: o.Writer,
	}
	n, err = fmt.Fprintf(w, format, a...)
	o.OutLine = w.n
	return
}
func (o *Outputer) PrintfLine(format string, a ...interface{}) (n int, err error) {
	o.outLine()
	n, err = o.Printf(format, a...)
	return
}
