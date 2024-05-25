package dep_tracer

import (
    "os"
    "fmt"
)

type OutputWriter interface {
    Println(args ...any)
    Print(args ...any)
}

func NewStdoutWriter() *StdoutWriter {
    return &StdoutWriter{}
}
type StdoutWriter struct {}
func (w *StdoutWriter) Println(args ...any) {
    fmt.Println(args...)
}
func (w *StdoutWriter) Print(args ...any) {
    fmt.Print(args...)
}

func NewFileWriter(path string) *FileWriter {
    f, err := os.Create(path)
    if err != nil {
        panic(err)
    }
    return &FileWriter{
        f: f,
    }
}
type FileWriter struct {
    f *os.File
}
func (w *FileWriter) Println(args ...any) {
    fmt.Fprintln(w.f, args...)
}
func (w *FileWriter) Print(args ...any) {
    fmt.Fprint(w.f, args...)
}
