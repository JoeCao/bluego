package s18

import (
	"bytes"
	"encoding/binary"
)

type Base struct {
	Head      uint8
	CommandId uint8
	Length    uint16
	Content   []byte
	CRC       uint8
	Tail      uint8
}

func NewBase() *Base {
	return &Base{
		Head:    0x68,
		Tail:    0x16,
		Content: []byte{},
	}
}

func (base *Base) ToFrame() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(base.Head)
	buf.WriteByte(base.CommandId)
	base.Length = uint16(len(base.Content))
	_ = binary.Write(buf, binary.LittleEndian, base.Length)
	buf.Write(base.Content)
	//求crc值
	var sum int32
	for _, x := range buf.Bytes() {
		sum = int32(x) + sum
	}
	//大端序，取int的最后8位bit值作为crc的校验
	//直接强转就是保留了最后的8位
	base.CRC = uint8(sum)
	buf.WriteByte(base.CRC)
	buf.WriteByte(base.Tail)
	return buf, nil
}
