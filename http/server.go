package http

import (
	"flag"
	"fmt"
	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/websocket"
	socketio "github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
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
	opts := engineio.Options{
		Transports: []transport.Transport{websocket.Default},
	}
	server, err := socketio.NewServer(&opts)
	if err != nil {
		log.Errorf("error")
		panic(err)
	}
	fileServer := http.FileServer(FileSystem{http.Dir(*directory)})
	http.Handle("/static/", http.StripPrefix(strings.TrimRight("/static/", "/"), fileServer))
	http.Handle("/", http.StripPrefix(strings.TrimRight("/", "/index.html"), fileServer))

	server.OnConnect("", func(s socketio.Conn) error {
		s.SetContext("")
		log.Infof("connected")
		return nil
	})
	server.OnEvent("/", "message", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})
	server.OnEvent("", "open", func(s socketio.Conn, msg string) {
		log.Errorf("error%s", msg)
		s.Emit("server_response", "start")
	})
	server.OnDisconnect("", func(s socketio.Conn, msg string) {
		fmt.Println("closed", msg)
	})
	server.OnError("", func(e error) {
		fmt.Println("meet error:", e)
	})
	http.Handle("/socket.io/", server)

	go server.Serve()
	defer server.Close()
	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
