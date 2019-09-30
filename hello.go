package main

import (
	"bluego/discovery"
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"rsc.io/quote"
)

type Battery struct {
	Head uint8
	ID   uint8
	A    uint8
	B    uint8
	CRC  uint8
	Tail uint8
}

func main() {
	fmt.Println(quote.Hello())
	braceletChan, exit, err := discovery.Run("hci0", false)
	if err != nil {
		log.Fatal("can not find")
	}

	defer exit()
	for dev := range braceletChan {
		pro := dev.Properties
		fmt.Printf("获取到设备名%s 设备地址%s ,设备类型%s, 强度%d \n",
			pro.Name, pro.Address, pro.AddressType, pro.RSSI)
		err = dev.Connect()
		if err != nil {
			log.Errorf("连接设备%s失败", pro.Name)
		}
		log.Infof("连接成功")
		chAll, err := dev.GetAllServicesAndUUID()
		for ch := range chAll {
			log.Info(ch)
		}

		rx, err := dev.GetCharByUUID("FFF0-FFF1")
		if err != nil {
			log.Errorf("rx error %s", err)
			dev.Disconnect()
			break
		}
		rx.StartNotify()
		tx, _ := dev.GetCharByUUID("FFF2")
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

			b, err := rx.ReadValue(nil)
			if err != nil {
				log.Errorf("got error while reading")
			}
			_ = fmt.Sprintf("%x", b)

		}()

	}
	select {}

}
