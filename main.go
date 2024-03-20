package main

import (
	"awesomeProject/cache"
	"awesomeProject/conn"
	"awesomeProject/handler"
	"awesomeProject/server"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}

	msgChan := make(chan handler.Msg, 1024)

	wg.Add(1)
	go func() {
		defer wg.Done()

	handlerLoop:
		for {
			select {
			case <-ctx.Done():
				log.Println("on handler context done:", ctx.Err())
				break handlerLoop

			case msg := <-msgChan:
				log.Println("on handler received:", string(msg.Data))
			}
		}
	}()

	opts := server.ServerOpts{
		ListenAddr: ":3000",
		MsgChan:    msgChan,
	}

	srv := server.New(opts, cache.NewCache())
	srv.Start()
	defer srv.Stop()

	go func() {
		log.Println("in client goroutine")
		cli, err := conn.Dial(":3000", time.Second*time.Duration(30))
		if err != nil {
			log.Fatalln("client conn error", err)
		}

		if err = cli.WriteLine([]byte("SET FOO BAR 1234")); err != nil {
			log.Fatalln("client writeline error", err)
		}

		log.Println("sent client message")
	}()

	s := <-sig

	cancel()

	log.Println("shutting down server by", s)

	wg.Wait()
}
