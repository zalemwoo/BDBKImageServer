package facecheck

import (
	"bytes"
	"encoding/binary"
	"log"
	"utils"
	"utils/crc16"
)

const (
	CMD_PREFIX  = "FR"
	CMD_VERSION = 0x01
	CMD_DETECT  = "DETECT"
)

func BuildMessage(filePath string) []byte {
	fileData, err := utils.FileReadBytes(filePath)
	if err != nil {
		log.Printf("ReadFile Error. filePath: [%s]. err: %v", filePath, err)
		return nil
	}

	bbuf := bytes.NewBuffer([]byte{})

	binary.Write(bbuf, binary.BigEndian, uint32(len(fileData)))
	dataLenBytes := make([]byte, 4, 4)
	copy(dataLenBytes, bbuf.Bytes())

	bbuf.Reset()

	cmdBytes := append([]byte(CMD_DETECT), 0x00)
	binary.Write(bbuf, binary.BigEndian, uint16(len(cmdBytes)))
	cmdLenBytes := make([]byte, 2, 2)
	copy(cmdLenBytes, bbuf.Bytes())

	dataBuffer := bytes.Buffer{}
	dataBuffer.Write([]byte(CMD_PREFIX))
	dataBuffer.WriteByte(CMD_VERSION)
	dataBuffer.Write(cmdLenBytes)
	dataBuffer.Write(cmdBytes)
	dataBuffer.Write(dataLenBytes)
	dataBuffer.Write(fileData)

	crc := crc16.ChecksumFaceRec(dataBuffer.Bytes())
	bbuf.Reset()
	binary.Write(bbuf, binary.BigEndian, crc)
	crcBytes := bbuf.Bytes()
	dataBuffer.Write(crcBytes)

	return dataBuffer.Bytes()
}

func ParseResult(message []byte) (totalLen int32, body string, isComplete bool) {
	messageLen := int32(len(message))
	cmdLenPart := message[3:5]
	b_buf := bytes.NewBuffer(cmdLenPart)
	var cmdLen int16
	binary.Read(b_buf, binary.BigEndian, &cmdLen)
	cmdLen32 := int32(cmdLen)

	dataLenPart := message[5+cmdLen32 : 5+cmdLen32+4]
	b_buf = bytes.NewBuffer(dataLenPart)
	var resultLen int32
	binary.Read(b_buf, binary.BigEndian, &resultLen)
	totalLen = 2 + 1 + 2 + cmdLen32 + 4 + resultLen + 2
	if messageLen >= totalLen {
		if resultLen == 0 || int(cmdLen32) != len("RESULT\x00") {
			return totalLen, "", true
		} else {
			return totalLen, string(message[5+cmdLen32+4 : 5+cmdLen32+4+resultLen-1]), true
		}
	}
	return totalLen, "", false
}
