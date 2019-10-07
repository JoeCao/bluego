package main

import (
	"bluego/discovery"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	log "github.com/sirupsen/logrus"
	"time"
)

type Battery struct {
	Head uint8
	ID   uint8
	A    uint8
	B    uint8
	CRC  uint8
	Tail uint8
}

type Bracelet struct {
	dev     *device.Device1
	ExitFun func()
}

func HandleBracelet() {
	braceletChan, exit, err := discovery.Run("hci0", false)
	if err != nil {
		log.Fatal("can not find")
	}
	var device1s []*device.Device1
	exif := func() {
		for _, dev := range device1s {
			log.Infof("disconnecting %s", dev.Properties.Name)
			_ = dev.Disconnect()
		}
		log.Info("stopping discovery")
		exit()
	}

	defer exif()
	for dev := range braceletChan {
		device1s = append(device1s, dev)
		InitDevice(dev)
	}
}

func InitDevice(dev *device.Device1) {
	pro := dev.Properties
	fmt.Printf("获取到设备名%s 设备地址%s ,设备类型%s, 强度%d \n",
		pro.Name, pro.Address, pro.AddressType, pro.RSSI)
	err := dev.Connect()
	if err != nil {
		log.Errorf("连接设备%s失败", pro.Name)
		panic("connection fail")
	}
	log.Infof("连接成功")
	//mapValue, err := dev.GetManufacturerData()
	//if err != nil {
	//	log.Errorf("can not get manufacture", err)
	//}
	//for k, v := range mapValue {
	//	fmt.Printf("key[%s] value[%s]\n", k, v)
	//}
	chAll, err := dev.GetCharacteristics()
	if err != nil {
		log.Errorf("can not get services", err)
	}
	log.Infof("属性长度 %d", len(chAll))
	for _, ch := range chAll {
		log.Infof("%s %s", ch.Properties.UUID, ch.Properties.Service)
	}
	rx, err := dev.GetCharByUUID("0000fff1-0000-1000-8000-00805f9b34fb")
	if err != nil {
		log.Errorf("rx error %s", err)
		panic("rx error")
	}
	//rx.StartNotify()
	tx, _ := dev.GetCharByUUID("0000fff2-0000-1000-8000-00805f9b34fb")
	var battery Battery
	battery.Head = 0x68
	battery.ID = 0x03
	battery.A = 0x00
	battery.B = 0x00
	battery.CRC = 0x6b
	battery.Tail = 0x16
	buf := &bytes.Buffer{}

	var _ = binary.Write(buf, binary.LittleEndian, battery)
	err = tx.WriteValue(buf.Bytes(), nil)
	if err != nil {
		log.Errorf("got error while write")
	}
	go func() {
		for {
			b, err := rx.ReadValue(nil)
			if err != nil {
				log.Errorf("got error while reading")
			}
			s := fmt.Sprintf("%x", b)
			log.Info(s)
			time.Sleep(time.Second)

		}

	}()
}
func main() {
	bracelet, err := device.NewDevice1("/org/bluez/hci0/dev_E2_C9_18_4F_8F_D9")
	if err != nil {
		log.Errorf(" connect bracelet fail")
	}
	var device1s []*device.Device1
	exif := func() {
		for _, dev := range device1s {
			log.Infof("disconnecting %s", dev.Properties.Name)
			_ = dev.Disconnect()
		}
		log.Info("stopping discovery")
	}
	defer exif()
	device1s = append(device1s, bracelet)
	InitDevice(bracelet)
	//HandleBracelet()
	select {}

}
