package s18

import (
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestNewResponse(t *testing.T) {
	s := "68860f00004b740000005800000035000000004916"
	data, err := hex.DecodeString(s)
	if err != nil {
		log.Infof("error happended")
	}
	r, err := NewResponse(&data)
	log.Info(r)
}
