// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crc16

// Predefined polynomials.
const (
	// Used by Bisync, Modbus, USB, ANSI X3.28, SIA DC-07, ...
	IBM = 0xA001

	// Used by X.25, V.41, HDLC FCS, XMODEM, Bluetooth, PACTOR, SD, ...
	CCITT = 0x8408

	// Used by SCSI
	SCSI = 0xEDD1

	// FaceRec
	FaceRec = 0x1021
)

// Table is a 256-word table representing the polynomial for efficient processing.
type Table [256]uint16

// IBMTable is the table for the IBM polynomial.
var IBMTable = makeTable(IBM)

// CCITTTable is the table for the CCITT polynomial.
var CCITTTable = makeTable(CCITT)

// SCSITable is the table for the SCSI polynomial.
var SCSITable = makeTable(SCSI)

var FaceTable = makeTable2(FaceRec)

// MakeTable returns the Table constructed from the specified polynomial.
func MakeTable(poly uint16) *Table {
	return makeTable(poly)
}

// makeTable returns the Table constructed from the specified polynomial.
func makeTable(poly uint16) *Table {
	t := new(Table)
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 == 1 {
				crc = (crc >> 1) ^ poly
			} else {
				crc >>= 1
			}
		}
		t[i] = crc
	}
	return t
}

func makeTable2(poly uint16) *Table {
	t := new(Table)
	for i := 0; i < 256; i++ {
		crc := uint16(i << 8)
		for j := 0; j < 8; j++ {
			if crc&0x8000 == 0x8000 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
		t[i] = crc
	}
	return t
}

func update(crc uint16, tab *Table, p []byte) uint16 {
	crc = ^crc
	for _, v := range p {
		crc = tab[byte(crc)^v] ^ (crc >> 8)
	}
	return ^crc
}

func update2(crc uint16, tab *Table, p []byte) uint16 {
	for _, v := range p {
		crcregister := byte(crc >> 8)
		crc <<= 8
		crc ^= tab[crcregister^v]
	}
	return crc
}

// Update returns the result of adding the bytes in p to the crc.
func Update(crc uint16, tab *Table, p []byte) uint16 {
	return update(crc, tab, p)
}

// Checksum returns the CRC-16 checksum of data
// using the polynomial represented by the Table.
func Checksum(data []byte, tab *Table) uint16 { return Update(0, tab, data) }

// ChecksumIBM returns the CRC-16 checksum of data
// using the IBM polynomial.
func ChecksumIBM(data []byte) uint16 { return update(0, IBMTable, data) }

// ChecksumCCITT returns the CRC-16 checksum of data
// using the CCITT polynomial.
func ChecksumCCITT(data []byte) uint16 { return update(0, CCITTTable, data) }

// ChecksumSCSI returns the CRC-16 checksum of data
// using the SCSI polynomial.
func ChecksumSCSI(data []byte) uint16 { return update(0, SCSITable, data) }

func ChecksumFaceRec(data []byte) uint16 { return update2(0, FaceTable, data) }
