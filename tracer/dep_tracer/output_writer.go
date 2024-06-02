package dep_tracer

import (
    "os"
    "fmt"
    "strings"
    "net/http"
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

func NewHttpWriter(url string) *HttpWriter {
    if !strings.HasPrefix(url, "http://") {
        panic("non http:// prefix")
    }
    addr := url[len("http://"):]

    res := &HttpWriter{
        data: []byte{},
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        w.Write([]byte(WebviewPageData))
    })

    http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
        w.Write(res.data)
    })

    http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
        res.data = []byte{}
    })

    go http.ListenAndServe(addr, nil)
    return res
}
type HttpWriter struct {
    data []byte
}
func (w *HttpWriter) Println(args ...any) {
    w.data = fmt.Appendln(w.data, args...)
}
func (w *HttpWriter) Print(args ...any) {
    w.data = fmt.Append(w.data, args...)
}