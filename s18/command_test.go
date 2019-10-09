package s18

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestBase_ToFrame(t *testing.T) {
	by := NewBase()
	by.CommandId = 0x03
	by.Content = []byte{}
	b, err := by.ToFrame()
	if err != nil {

	}
	s := fmt.Sprintf("%x", b)
	logrus.Infof(s)

}

func TestBase_ToFrame2(t *testing.T) {
	x := &bytes.Buffer{}
	x.WriteByte(0x00)
	x.Write([]byte("13656898745张三"))
	base := NewBase()
	base.CommandId = 0x08
	base.Content = x.Bytes()
	b, err := base.ToFrame()
	if err != nil {

	}
	s := fmt.Sprintf("%x", b)
	logrus.Infof(s)
}
