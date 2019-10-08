package main

import (
	"bluego/discovery"
	"bluego/s18"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

func HandleBracelet(ch chan os.Signal) {
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
			bracelet, _ := s18.RBracelet(dev, ch)
			bracelets = append(bracelets, bracelet)

		case sig := <-ch:
			log.Infof("收到退出的消息 %s", sig)
			goto end

		}
	}
end:
}

func main() {
	var bracelets []*s18.Bracelet
	exif := func() {
		for _, b := range bracelets {
			log.Infof("disconnecting %s", b.Name)
			_ = b.Disconnect()
		}
	}
	defer exif()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	bracelet, _ := s18.OpenBracelet("/org/bluez/hci0/dev_E2_C9_18_4F_8F_D9", c)
	bracelets = append(bracelets, bracelet)
	_ = bracelet.GetBattery()
	time.Sleep(2 * time.Second)

	//HandleBracelet(c)
	_ = bracelet.GetVersion()
	time.Sleep(2 * time.Second)
	_ = bracelet.GetHeartBeat()
	time.Sleep(4 * time.Second)
}
