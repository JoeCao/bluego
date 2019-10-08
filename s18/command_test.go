package s18

import (
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
