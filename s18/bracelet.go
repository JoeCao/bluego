package s18

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/godbus/dbus"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	"github.com/muka/go-bluetooth/bluez/profile/gatt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"time"
)

type Bracelet struct {
	Name      string
	dev       *device.Device1
	rx        *gatt.GattCharacteristic1
	tx        *gatt.GattCharacteristic1
	retCh     chan *[]byte
	connQueue chan string
}

func RBracelet(device *device.Device1, ch chan string) (bracelet *Bracelet, err error) {
	br := &Bracelet{
		dev:       device,
		Name:      device.Properties.Name,
		retCh:     make(chan *[]byte),
		connQueue: make(chan string),
	}
	_ = br.InitBracelet(ch)
	return br, nil
}

func OpenBracelet(path dbus.ObjectPath, ch chan string) (bracelet *Bracelet, err error) {
	d, err := device.NewDevice1(path)
	if err != nil {
		panic(err)
	}
	bracelet, err = RBracelet(d, ch)
	return bracelet, nil
}
func (bracelet *Bracelet) InitBracelet(ch chan string) (err error) {
	dev := bracelet.dev
	pro := dev.Properties
	fmt.Printf("获取到设备名%s 设备地址%s ,设备类型%s, 强度%d \n",
		pro.Name, pro.Address, pro.AddressType, pro.RSSI)
	err = dev.Connect()
	if err != nil {
		log.Errorf("连接设备%s失败", pro.Name)
		return err
	}
	log.Infof("连接成功%s", bracelet.Name)
	go func() {
		for x := range bracelet.connQueue {
			log.Warnf("感知到断开，开始重连 %s", x)
			f, _ := bracelet.dev.GetConnected()
			if !f {
				log.Info("开始重连")
				_ = bracelet.dev.Disconnect()
				log.Info("断开成功")
				err = bracelet.dev.Connect()
				if err != nil {
					log.Warn("重连失败，等待重试")
				}
				log.Info("重连成功")
			} else {
				log.Info("仍然是连接状态")
			}

		}
	}()
	//暂停1000ms等消息返回
	time.Sleep(1000 * time.Millisecond)

	chAll, err := dev.GetCharacteristics()
	if err != nil {
		log.Errorf("can not get services %s", err)
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
				//处理分包的问题 0x68和0x16是每帧头尾标志
				if b1[0] == 0x68 && b1[len(b1)-1] == 0x16 {
					log.Debugf("获取的返回字节为%x", b1)
					bracelet.retCh <- &b1
				} else if b1[0] == 0x68 {
					_ = binary.Write(buffer, binary.BigEndian, b1)
				} else if b1[len(b1)-1] == 0x16 {
					_ = binary.Write(buffer, binary.BigEndian, b1)
					tmp := buffer.Bytes()
					log.Debugf("获取返回字节为%x", tmp)
					bracelet.retCh <- &tmp
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
	bs, _ := base.ToFrame()
	s := fmt.Sprintf("%x", bs)
	log.Debugf("写入的字符byte为 %s", s)
	err = bracelet.tx.WriteValue(bs.Bytes(), nil)
	if err != nil {
		log.Errorf("got error while write")
		bracelet.connQueue <- "error"
		return err
	}
	return nil
}

func (bracelet *Bracelet) GetBattery() (capacity uint8, err error) {
	base := NewBase()
	base.CommandId = 0x03
	_ = bracelet.write(base)
	select {
	case b1 := <-bracelet.retCh:
		b2 := *b1
		return b2[4], nil
	}
}

func (bracelet *Bracelet) GetVersion() (v interface{}, err error) {
	base := NewBase()
	base.CommandId = 0x07
	_ = bracelet.write(base)
	callback := func(b1 *[]byte) (interface{}, error) {
		b2 := *b1
		slice := b2[4:8]
		return fmt.Sprintf("%x", slice), nil
	}
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) StringTracing() (ok interface{}, err error) {
	base := NewBase()
	base.CommandId = 0x06
	base.Content = []byte{0x01}
	callback := func(b1 *[]byte) (interface{}, error) {
		return fmt.Sprintf("%x", *b1), nil
	}
	_ = bracelet.write(base)
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) StopTracing() (ok interface{}, err error) {
	base := NewBase()
	base.CommandId = 0x06
	base.Content = []byte{0x02}
	callback := func(b1 *[]byte) (interface{}, error) {
		return fmt.Sprintf("%x", *b1), nil
	}
	_ = bracelet.write(base)
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) Tracing() (h interface{}, err error) {
	base := NewBase()
	base.CommandId = 0x06
	base.Content = []byte{0x00}
	_ = bracelet.write(base)
	callback := func(b1 *[]byte) (interface{}, error) {
		resp, err := NewResponse(b1)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) Reset() (ok interface{}, err error) {
	base := NewBase()
	base.CommandId = 0x11
	base.Content = []byte{0x01}
	_ = bracelet.write(base)
	callback := func(b1 *[]byte) (interface{}, error) {
		return fmt.Sprintf("%x", *b1), nil
	}
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) Notification(content string) (ok interface{}, err error) {
	x := &bytes.Buffer{}
	x.WriteByte(0x00)
	x.Write([]byte(content))
	base := NewBase()
	base.CommandId = 0x08
	base.Content = x.Bytes()
	_ = bracelet.write(base)
	callback := func(b1 *[]byte) (interface{}, error) {
		return fmt.Sprintf("%x", *b1), nil
	}
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) CallNoti(content string) (ok interface{}, err error) {
	x := &bytes.Buffer{}
	x.Write([]byte(content))
	base := NewBase()
	base.CommandId = 0x01
	base.Content = x.Bytes()
	_ = bracelet.write(base)
	callback := func(b1 *[]byte) (interface{}, error) {
		return fmt.Sprintf("%x", *b1), nil
	}
	return bracelet.getRet(callback)
}

func (bracelet *Bracelet) getRet(callback func(*[]byte) (interface{}, error)) (ret interface{}, err error) {
	select {
	case b1 := <-bracelet.retCh:
		return callback(b1)
	case <-time.After(3 * time.Second):
		return nil, errors.New("time out")
	}
}
func TestOperration() {
	var bracelets []*Bracelet
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

	bracelet, _ := OpenBracelet("/org/bluez/hci0/dev_E2_C9_18_4F_8F_D9", ch)
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
	log.Info(bracelet.StringTracing())
	for {
		select {
		case <-time.After(time.Second * 1):
			resp, _ := bracelet.Tracing()
			log.Infof("心跳为%v", resp)
		case sig := <-c:
			log.Info("收到操作系统的消息%s", sig)
			goto end
		}
	}
end:
	log.Info(bracelet.StopTracing())
	ch <- "退出"
}
