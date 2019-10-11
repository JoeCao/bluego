package http

import (
	"bluego/discovery"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Init() {

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
			m := map[string]string{
				"addr":              l.Properties.Address,
				"CompleteLocalName": l.Properties.Name,
				"addrType":          l.Properties.AddressType,
			}
			devs = append(devs, m)

		}
		c.JSON(200, devs)
	})
	r.Run(":8000")
}
