package http

import (
	"bluego/discovery"
	"bluego/s18"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/godbus/dbus"
	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/websocket"
	socketio "github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var SocketioServer *socketio.Server
var ConnectedMap = make(map[string]*s18.Bracelet)

func respStr(resp string) string {
	m := make(map[string]string)
	m["data"] = resp
	s, _ := json.Marshal(m)
	return string(s)
}

func Init() {
	var err error
	opts := engineio.Options{
		Transports: []transport.Transport{websocket.Default},
	}
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
	SocketioServer.OnEvent("", "start", func(s socketio.Conn, msg string) {
		log.Infof("收到消息%s", msg)
		if ConnectedMap[msg] != nil {
			brace := ConnectedMap[msg]
			s.Emit("server_response", respStr("实时心率监测开始"))
			_, _ = brace.StartTracing()
			go func() {
				for {
					select {
					case <-time.After(time.Second * 1):
						resp, err := brace.Tracing()
						if err != nil {
							log.Warn("获取空对象")
							continue
						}
						hb := resp.(s18.HeartBeatResponse)
						b, _ := json.Marshal(map[string]string{
							"heart_beat":    fmt.Sprintf("%d", hb.HeartBeat),
							"peace_count":   fmt.Sprintf("%d", hb.PeaceCount),
							"meter_count":   fmt.Sprintf("%d", hb.MeterCount),
							"calorie":       fmt.Sprintf("%d", hb.Calorie),
							"data":          "运动中",
							"bracelet_name": brace.Name,
						})
						s.Emit("server_response", string(b))
						log.Infof("%v", string(b))
					case sig := <-brace.StopTraceChannel:
						log.Infof("收到结束监控的消息%s", sig)
						goto end
					}
				}
			end:
				log.Infof("%s的实时心率监控结束", brace.Name)
			}()
			s.Emit("command_response", respStr("实时心率检测开始"))

		}
	})
	SocketioServer.OnEvent("", "stop", func(s socketio.Conn, msg string) {
		log.Infof("收到消息%s", msg)
		if ConnectedMap[msg] != nil {
			s.Emit("server_response", respStr("实时心率监测结束"))
			brace := ConnectedMap[msg]
			brace.StopTraceChannel <- "stop"
			_, _ = brace.StopTracing()
			s.Emit("command_response", respStr("实时心率监测结束"))
		}
	})
	SocketioServer.OnEvent("", "open", func(s socketio.Conn, msg string) {
		log.Infof("get path is %s", msg)
		if ConnectedMap[msg] != nil {
			log.Infof("已经存在的设备%s", msg)
			return
		}
		s.Emit("server_response", respStr("开始连接手环"))
		bracelet, err := s18.OpenBracelet(dbus.ObjectPath(msg))
		if err != nil {
			log.Errorf("连接%s失败", msg)
			return
		}
		ConnectedMap[bracelet.Name] = bracelet
		s.Emit("command_response", respStr("手环已经连接"))
	})
	SocketioServer.OnEvent("", "message", func(s socketio.Conn, msg string) {
		log.Infof("notice:%s", msg)
		s.Emit("command_response", "reply to "+msg)
	})
	go SocketioServer.Serve()
	defer SocketioServer.Close()
	r := gin.Default()
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
	r.GET("/main", func(c *gin.Context) {
		x := ""
		c.HTML(http.StatusOK, "main.html", x)
	})
	r.GET("/get_base_data", func(c *gin.Context) {
		var rets = make([]*map[string]string, 0)
		for _, a := range ConnectedMap {
			m := map[string]string{
				"localName": a.Name,
				"address":   a.Address,
				"addrType":  a.AddressType,
				"statusStr": a.Status,
			}
			rets = append(rets, &m)
		}
		c.JSON(200, rets)
	})
	r.GET("/socket.io/", gin.WrapH(SocketioServer))
	r.POST("/socket.io/", gin.WrapH(SocketioServer))
	r.Handle("WS", "/socket.io/", gin.WrapH(SocketioServer))
	r.Handle("WSS", "/socket.io/", gin.WrapH(SocketioServer))

	r.Run(":8000")
}
