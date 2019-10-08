package s18

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type HeartBeatResponse struct {
	Flag       byte
	HeartBeat  byte
	PeaceCount uint32
	MeterCount uint32
	Calorie    uint32
	PeaceSpeed byte
}

func NewResponse(btr *[]byte) (h HeartBeatResponse, err error) {
	b := *btr
	h = HeartBeatResponse{}

	if len(b) > 6 {
		l := b[2]
		data := b[4 : 4+l]
		buffer := &bytes.Buffer{}
		buffer.Write(data)
		h.Flag, _ = buffer.ReadByte()
		h.HeartBeat, _ = buffer.ReadByte()
		_ = binary.Read(buffer, binary.LittleEndian, &h.PeaceCount)
		_ = binary.Read(buffer, binary.LittleEndian, &h.MeterCount)
		_ = binary.Read(buffer, binary.LittleEndian, &h.Calorie)
		h.PeaceSpeed, _ = buffer.ReadByte()
		return h, nil

	} else {
		return h, errors.New("not suitabe")
	}
}
