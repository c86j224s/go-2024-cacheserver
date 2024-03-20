package conn

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"os"
	"time"
)

type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
	buf    []byte
}

func New(conn net.Conn) *Conn {
	return &Conn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		buf:    []byte{},
	}
}

func Dial(addr string, timeout time.Duration) (*Conn, error) {
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	return &Conn{
		conn:   c,
		reader: bufio.NewReader(c),
		buf:    []byte{},
	}, nil
}

func (c *Conn) Close() {
	if err := c.conn.Close(); err != nil {
		log.Println("conn close error:", err)
	}
}

func (c *Conn) ReadLine(ctx context.Context) ([]byte, error) {
	const deadLine = time.Second

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default:
			if err := c.conn.SetReadDeadline(time.Now().Add(deadLine)); err != nil {
				return nil, err
			}

			ln, remain, err := c.reader.ReadLine()
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}

				return nil, err
			}

			c.buf = append(c.buf, ln...)

			if !remain {
				ret := c.buf
				c.buf = []byte{}

				return ret, nil
			}
		}
	}
}

func (c *Conn) WriteLine(data []byte) error {
	data = append(data, '\n')
	_, err := c.conn.Write(data)
	return err
}
