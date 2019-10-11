package http

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

// FileSystem custom file system handler
type FileSystem struct {
	fs http.FileSystem
}

// Open opens file
func (fs FileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}

func Serve() {
	port := flag.String("p", "3000", "port to serve on")
	directory := flag.String("d", "static", "the directory of static file to host")
	flag.Parse()

	fileServer := http.FileServer(FileSystem{http.Dir(*directory)})
	http.Handle("/static/", http.StripPrefix(strings.TrimRight("/static/", "/"), fileServer))
	http.Handle("/", http.StripPrefix(strings.TrimRight("/", "/index.html"), fileServer))
	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
