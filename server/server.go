package server

import (
	"awesomeProject/cache"
	"awesomeProject/conn"
	"awesomeProject/handler"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type ServerOpts struct {
	ListenAddr string

	MsgChan chan handler.Msg
}

type Server struct {
	ServerOpts

	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	cache cache.Cacher

	lock     sync.Mutex
	listener net.Listener
}

func New(opts ServerOpts, c cache.Cacher) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		ServerOpts: opts,
		wg:         &sync.WaitGroup{},
		ctx:        ctx,
		cancel:     cancel,
		cache:      c,
		listener:   nil,
	}
}

func (s *Server) Stop() error {
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}
	s.cancel()
	s.wg.Wait()

	return nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error, %s", err)
	}

	s.listener = ln

	log.Println("server listening on", s.ListenAddr)

	go func() {
		for {
			c, err := s.listener.Accept()
			if err != nil {
				log.Println("accept error", err)
				break
			}

			s.wg.Add(1)
			go func() {
				defer s.wg.Done()

				s.handleConn(c)
			}()
		}
	}()

	return nil
}

func (s *Server) handleConn(netConn net.Conn) {
	c := conn.New(netConn)

	for {
		data, err := c.ReadLine(s.ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("handleConn connection eof")
				return
			}

			log.Println("handleConn readline error:", err)
			return
		}

		if data == nil {
			log.Println("handleConn connection closed:", err)
			return
		}

		msg := string(data)
		log.Println("handleConn received:", msg)

		if s.MsgChan != nil {
			s.MsgChan <- handler.Msg{
				Ctx:  s.ctx,
				Wg:   s.wg,
				C:    c,
				Data: data,
			}
		}
	}
}
