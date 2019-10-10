package main

import (
	"bluego/discovery"
	"bluego/s18"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
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
	list, _ := discovery.RunWithin("hci0", 10)
	for _, l := range list {
		log.Infof("name:%s rssi:%d", l.Properties.Name, l.Properties.RSSI)
	}
	time.Sleep(10 * time.Second)
	var bracelets []*s18.Bracelet
	exif := func() {
		for _, b := range bracelets {
			log.Infof("disconnecting %s", b.Name)
			_ = b.Disconnect()
		}
	}
	defer exif()

	c := make(chan os.Signal, 1)
	ch := make(chan string)
	signal.Notify(c, os.Interrupt)

	bracelet, _ := s18.OpenBracelet("/org/bluez/hci0/dev_E2_C9_18_4F_8F_D9", ch)
	bracelets = append(bracelets, bracelet)
	capacity, _ := bracelet.GetBattery()
	log.Infof("剩余电量为%d", capacity)
	time.Sleep(2 * time.Second)
	//HandleBracelet(c)
	v, _ := bracelet.GetVersion()
	log.Infof("版本号为%s", v)
	time.Sleep(2 * time.Second)
	log.Info(bracelet.Notification("曹祖鹏"))
	log.Info(bracelet.Reset())
	time.Sleep(100 * time.Millisecond)
	log.Info(bracelet.StartHeartBeat())
	for {
		select {
		case <-time.After(time.Second * 1):
			resp, _ := bracelet.GetHeartBeat()
			log.Infof("心跳为%v", resp)
		case sig := <-c:
			log.Info("收到操作系统的消息%s", sig)
			goto end
		}
	}
end:
	log.Info(bracelet.StopHeartBeat())
	ch <- "退出"
}
