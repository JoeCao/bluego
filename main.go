package main

import (
	"bluego/discovery"
	"bluego/http"
	"bluego/s18"
	log "github.com/sirupsen/logrus"
)

func HandleBracelet(ch chan string) {
	braceletChan, exit, err := discovery.Run("hci0", false)
	if err != nil {
		log.Fatal("can not find")
	}
	var bracelets []*s18.Bracelet
	exif := func() {
		for _, b := range bracelets {
			log.Infof("disconnecting %s", b.Name)
			_ = b.Disconnect()
		}
		log.Info("stopping discovery")
		exit()
	}
	defer exif()
	for {
		select {
		case dev := <-braceletChan:
			bracelet, _ := s18.RBracelet(dev)
			bracelets = append(bracelets, bracelet)

		case sig := <-ch:
			log.Infof("收到退出的消息 %s", sig)
			goto end

		}
	}
end:
}

func main() {
	http.Init()
	//log.SetLevel(log.DebugLevel)
	//s18.TestOperation()
}
