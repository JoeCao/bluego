package http

import (
	"bluego/discovery"
	"bluego/s18"
	"fmt"
	"github.com/gin-gonic/gin"
	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/websocket"
	socketio "github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var SocketioServer *socketio.Server
var list []*s18.Bracelet

func socketHandler(c *gin.Context) {
	SocketioServer.ServeHTTP(c.Writer, c.Request)

}

func Init() {
	var err error
	opts := engineio.Options{
		Transports: []transport.Transport{websocket.Default},
	}
	//server, _ := socketio.NewServer(nil)
	SocketioServer, err = socketio.NewServer(&opts)
	if err != nil {
		log.Errorf("error")
		panic(err)
	}
	SocketioServer.OnConnect("", func(s socketio.Conn) error {
		s.SetContext("")
		fmt.Println("connected")
		return nil
	})
	SocketioServer.OnConnect("", func(s socketio.Conn) error {
		s.SetContext("")
		log.Infof("connected: %s", s.ID())
		return nil
	})
	SocketioServer.OnEvent("", "message", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})
	SocketioServer.OnEvent("", "open", func(s socketio.Conn, msg string) {
		log.Errorf("error%s", msg)
		s.Emit("server_response", "start")
	})
	SocketioServer.OnEvent("", "open", func(s socketio.Conn, msg string) {
		log.Errorf("error%s", msg)
		s.Emit("server_response", "start")
	})
	r := gin.Default()
	//r.Static("/static", "./static")
	r.StaticFS("/static", gin.Dir("./static", true))

	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		x := ""
		c.HTML(http.StatusOK, "index.html", x)
	})

	r.GET("/scan", func(c *gin.Context) {
		list, _ := discovery.RunWithin("hci0", 10)
		var devs []map[string]string
		for _, l := range list {
			devs = append(devs, l)

		}
		c.JSON(200, devs)
	})
	r.GET("/get_base_data", func(c *gin.Context) {
		var retlist []*map[string]string
		for a := 0; a < 1; a++ {
			m := map[string]string{
				"localName": "234234234",
				"addrType":  "random",
				"statusStr": "连接中",
			}
			retlist = append(retlist, &m)

		}
		c.JSON(200, retlist)
	})
	r.GET("/socket.io/", socketHandler)
	r.POST("/socket.io/", socketHandler)
	r.Handle("WS", "/socket.io/", socketHandler)
	r.Handle("WSS", "/socket.io/", socketHandler)
	//r.GET("/socket.io/", gin.WrapH(server))

	r.Run(":8000")
}
