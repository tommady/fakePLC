package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	bigEndian    = "BigEndian"
	littleEndian = "LittleEndian"
)

const (
	statusOK = iota + 1
	statusAuthFailed
	statusProcessFailed
	statusInternalServerError
)

const (
	responseCmd = iota + 1
	authCmd
	barcodesCmd
	hearbeatCmd
)

type packet struct {
	endian binary.ByteOrder
}

func newPacket(endian string) (*packet, error) {
	p := new(packet)

	switch endian {
	case bigEndian:
		p.endian = binary.BigEndian
	case littleEndian:
		p.endian = binary.LittleEndian
	default:
		return nil, errors.New("unknown endian type: " + endian)
	}

	return p, nil
}

func (p *packet) unpackCmdHeader(data *bufio.Reader) (int, error) {
	b, err := unpack(data)
	if err != nil {
		return 0, err
	}
	return int(p.endian.Uint16(b.Bytes())), nil
}

func unpack(data *bufio.Reader) (*bytes.Buffer, error) {
	b, err := data.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// trim the request string, ReadString does not strip any newlines.
	b = bytes.TrimRight(b, "\r\n ")
	return bytes.NewBuffer(b), nil
}

func (p *packet) writeHeartbeat(writer *bufio.Writer) {
	cmd := make([]byte, 2)
	p.endian.PutUint16(cmd, hearbeatCmd)
	writer.Write(cmd)
	writer.WriteByte('\r')
	writer.WriteByte('\n')
	writer.Flush()
}

type response struct {
	status int
	msg    string
}

type responsePacket struct {
	Scode uint16
	Msg   [60]byte
}

func (p *packet) unpackResponse(data *bufio.Reader) (*response, error) {
	b, err := unpack(data)
	if err != nil {
		return nil, err
	}
	res := new(responsePacket)
	if err := binary.Read(b, p.endian, res); err != nil {
		return nil, err
	}

	return &response{
		status: int(res.Scode),
		msg:    string(res.Msg[:]),
	}, nil
}
