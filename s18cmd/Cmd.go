package s18cmd

import "bytes"

type Base struct {
	Head      uint8
	CommandId uint8
	Length    uint16
	Content   *bytes.Buffer
	CRC       uint8
	Tail      uint8
}
