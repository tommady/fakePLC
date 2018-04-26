package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"strconv"
	"time"
)

type client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	packet *packet
	start  time.Time
	done   chan struct{}
}

func newClient(addr, ssid, endian string, localPort int) (*client, error) {
	d := net.Dialer{
		LocalAddr: &net.TCPAddr{Port: localPort},
	}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	p, err := newPacket(endian)
	if err != nil {
		return nil, err
	}

	c := &client{
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		packet: p,
		done:   make(chan struct{}),
		conn:   conn,
	}

	if err = c.auth(ssid); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *client) close() error {
	select {
	case <-c.done:
		// already closed
	default:
		close(c.done)
		c.conn.Close()
	}
	return nil
}

func (c *client) receiveResponse() (*response, error) {
	cmd, err := c.packet.unpackCmdHeader(c.reader)
	if err != nil {
		return nil, err
	}

	switch cmd {
	case responseCmd:
		res, err := c.packet.unpackResponse(c.reader)
		if err != nil {
			return nil, err
		}
		return res, nil

	default:
		return nil, errors.New("receive a not response packet: " + strconv.Itoa(cmd))
	}
}

func (c *client) handlingLongTerm() {
	go func() {
		defer c.close()
		for {
			cmd, err := c.packet.unpackCmdHeader(c.reader)
			if err != nil {
				log.Printf("read failed:%v", err)
				break
			}

			switch cmd {
			case responseCmd:
				res, err := c.packet.unpackResponse(c.reader)
				if err != nil {
					log.Printf("receive response failed:%v", err)
				}
				log.Printf("basket[%s] took %s", res.msg, time.Since(c.start))
				switch res.status {
				case statusOK:
					log.Printf("basket[%s] success", res.msg)
				case statusProcessFailed:
					log.Printf("basket[%s] process failed", res.msg)
				}

			case hearbeatCmd:
				c.packet.writeHeartbeat(c.writer)
				log.Printf("handled heartbeat")
			default:
				log.Printf("recieve unknown cmd:%d", cmd)
				break
			}
		}
	}()
}

func (c *client) auth(ssid string) error {
	binary.Write(c.writer, binary.BigEndian, uint16(authCmd))
	binary.Write(c.writer, binary.BigEndian, [2]byte{'\r', '\n'})

	sid := [32]byte{}
	copy(sid[:], ssid)
	binary.Write(c.writer, binary.BigEndian, sid)
	binary.Write(c.writer, binary.BigEndian, [2]byte{'\r', '\n'})
	if err := c.writer.Flush(); err != nil {
		return err
	}

	cmd, err := c.packet.unpackCmdHeader(c.reader)
	if err != nil {
		return err
	}
	if cmd != responseCmd {
		return errors.New("auth recieve not a response command")
	}

	res, err := c.packet.unpackResponse(c.reader)
	if err != nil {
		return err
	}

	if res.status != statusOK {
		return errors.New("auth recieve status not OK")
	}

	return nil
}

func (c *client) purchase(basketID string, barcodes []string) error {
	c.start = time.Now()
	binary.Write(c.writer, c.packet.endian, uint16(3))
	binary.Write(c.writer, c.packet.endian, [2]byte{'\r', '\n'})
	binary.Write(c.writer, c.packet.endian, uint16(len(basketID)))
	binary.Write(c.writer, c.packet.endian, uint32(len(barcodes)))
	bsid := make([]byte, 58)
	copy(bsid[:], basketID)
	binary.Write(c.writer, c.packet.endian, bsid)
	bcode := make([]byte, 30)
	for _, barcode := range barcodes {
		binary.Write(c.writer, c.packet.endian, uint16(len(barcode)))
		copy(bcode[:], barcode)
		binary.Write(c.writer, c.packet.endian, bcode)
	}
	binary.Write(c.writer, c.packet.endian, [2]byte{'\r', '\n'})
	return c.writer.Flush()
}
