package handler

import (
	"awesomeProject/conn"
	"context"
	"sync"
)

type Msg struct {
	Ctx  context.Context
	Wg   *sync.WaitGroup
	C    *conn.Conn
	Data []byte
}
