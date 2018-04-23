package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log"
	"net"
)

type client struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	packet *packet
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

	go c.handler()

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

func (c *client) handler() {
	defer c.close()

	go func() {
		for {
			cmd, err := c.packet.unpackCmdHeader(c.reader)
			if err != nil {
				log.Printf("[PLC] read failed:%v", err)
				break
			}

			switch cmd {
			case responseCmd:
				if res, err := c.packet.unpackResponse(c.reader); err != nil {
					log.Printf("[PLC] unpack response failed:%v", err)
				} else {
					log.Printf("[PLC] response status%v, msg:%v", res.status, res.msg)
				}

			case hearbeatCmd:
				c.packet.writeHeartbeat(c.writer)
			default:
				log.Printf("[PLC] recieve unknown cmd:%d", cmd)
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
	c.writer.Flush()

	cmd, err := c.packet.unpackCmdHeader(c.reader)
	if err != nil {
		return err
	}
	if cmd != responseCmd {
		return errors.New("recieve not response command: " + string(cmd))
	}

	res, err := c.packet.unpackResponse(c.reader)
	if err != nil {
		return err
	}

	if res.status != statusOK {
		return errors.New("recieve auth status not OK: " + string(res.status) + ", " + res.msg)
	}

	return nil
}
