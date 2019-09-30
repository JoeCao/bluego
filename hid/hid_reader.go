package hid

import (
	"fmt"
	"github.com/karalabe/hid"
	log "github.com/sirupsen/logrus"
)

func Enumerate() {
	var infos = hid.Enumerate(0, 0)
	for _, deviceInfo := range infos {
		log.Infof("VendorID:%d, ProductID:%d, Manufacturer:%s \n",
			deviceInfo.VendorID, deviceInfo.ProductID, deviceInfo.Manufacturer)
	}
	var devices = hid.Enumerate(65535, 53)
	log.Infof("%d", len(devices))
	if len(devices) >= 1 {
		deviceInfo := devices[0]
		log.Infof("VendorID:%d, ProductID:%d, Manufacturer:%s \n",
			deviceInfo.VendorID, deviceInfo.ProductID, deviceInfo.Manufacturer)
		device, err := deviceInfo.Open()
		if err != nil {
			log.Errorf("can not open device !")
		}
		go func() {
			var b []byte
			length, err := device.Read(b)
			if err != nil {
				log.Errorf("read error")
			}
			if length >= 0 {
				for by := range b {
					fmt.Print(by)
				}
			}
		}()
		select {}
	} else {
		log.Print("can not find device ")
	}
}
