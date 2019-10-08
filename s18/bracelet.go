package s18

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/godbus/dbus"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type Bracelet struct {
	Name string
	dev  *device.Device1
	rx   *gatt.GattCharacteristic1
	tx   *gatt.GattCharacteristic1
}

func RBracelet(device *device.Device1, ch chan os.Signal) (bracelet *Bracelet, err error) {
	br := &Bracelet{
		dev:  device,
		Name: device.Properties.Name,
	}
	br.InitBracelet(ch)
	return br, nil
}

func OpenBracelet(path dbus.ObjectPath, ch chan os.Signal) (bracelet *Bracelet, err error) {
	d, err := device.NewDevice1(path)
	if err != nil {
		panic(err)
	}
	bracelet, err = RBracelet(d, ch)
	return bracelet, nil
}
func (bracelet *Bracelet) InitBracelet(ch chan os.Signal) (err error) {
	dev := bracelet.dev
	pro := dev.Properties
	fmt.Printf("获取到设备名%s 设备地址%s ,设备类型%s, 强度%d \n",
		pro.Name, pro.Address, pro.AddressType, pro.RSSI)
	err = dev.Connect()
	if err != nil {
		log.Errorf("连接设备%s失败", pro.Name)
		return err
	}
	log.Infof("连接成功")
	//暂停1000ms等消息返回
	time.Sleep(1000 * time.Millisecond)

	chAll, err := dev.GetCharacteristics()
	if err != nil {
		log.Errorf("can not get services", err)
	}
	log.Infof("Characteristics 属性长度 %d", len(chAll))
	for _, ch := range chAll {
		log.Infof("%s %s", ch.Properties.UUID, ch.Properties.Service)
	}
	//打开读通道，读写通道是异步的、分离的
	bracelet.rx, err = dev.GetCharByUUID("0000fff1-0000-1000-8000-00805f9b34fb")
	if err != nil {
		log.Errorf("rx error %s", err)
		return err
	}
	_ = bracelet.rx.StartNotify()
	propsChanged, err := bracelet.rx.WatchProperties()
	if err != nil {
		return err
	}

	go func() {
		buffer := &bytes.Buffer{}
		for {
			select {
			case prop := <-propsChanged:
				if prop == nil {
					return
				}
				if prop.Name != "Value" {
					return
				}
				b1 := prop.Value.([]byte)
				if b1[0] == 0x68 && b1[len(b1)-1] == 0x16 {
					log.Infof("获取的返回字节为%x", b1)
				} else if b1[0] == 0x68 {
					_ = binary.Write(buffer, binary.BigEndian, b1)
				} else if b1[len(b1)-1] == 0x16 {
					_ = binary.Write(buffer, binary.BigEndian, b1)
					log.Infof("获取返回字节为%x", buffer.Bytes())
					buffer.Reset()
				}
			case sig := <-ch:
				log.Infof("收到退出的消息 %s", sig)
				goto end
			}
		}
	end:
	}()
	//打开写通道
	bracelet.tx, err = dev.GetCharByUUID("0000fff2-0000-1000-8000-00805f9b34fb")
	if err != nil {
		log.Errorf("tx error %s", err)
		return err
	}
	return nil

}

func (bracelet *Bracelet) Disconnect() error {
	return bracelet.dev.Disconnect()
}

func (bracelet *Bracelet) write(base *Base) (err error) {
	byte, _ := base.ToFrame()
	s := fmt.Sprintf("%x", byte)
	log.Infof("写入的字符byte为 %s", s)
	err = bracelet.tx.WriteValue(byte.Bytes(), nil)
	if err != nil {
		log.Errorf("got error while write")
		return err
	}
	return nil
}

func (bracelet *Bracelet) GetBattery() (err error) {
	base := NewBase()
	base.CommandId = 0x03
	return bracelet.write(base)
}

func (bracelet *Bracelet) GetVersion() (err error) {
	base := NewBase()
	base.CommandId = 0x07
	return bracelet.write(base)
}

func (bracelet *Bracelet) StartHeartBeat() (err error) {
	base := NewBase()
	base.CommandId = 0x16
	base.Content = []byte{0x01}
	return bracelet.write(base)
}

func (bracelet *Bracelet) GetHeartBeat() (err error) {
	base := NewBase()
	base.CommandId = 0x06
	base.Content = []byte{0x00}
	return bracelet.write(base)
}
