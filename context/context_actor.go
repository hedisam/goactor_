package context

import (
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"time"
)

type actorContext struct {
	pid  pid.PID
	args []interface{}
	done chan struct{}
}

func NewContext(pid pid.PID, args []interface{}) Context {
	return &actorContext{
		pid:  pid,
		args: args,
		done: make(chan struct{}),
	}
}

func (ctx *actorContext) Args() []interface{} {
	return ctx.args
}

func (ctx *actorContext) Recv(handler mailbox.MessageHandler) {
	ctx.pid.Mailbox().Receive(handler)
}

func (ctx *actorContext) RecvWithTimeout(d time.Duration, handler mailbox.MessageHandler) {
	if d < 1 {
		ctx.pid.Mailbox().Receive(handler)
		return
	}
	ctx.pid.Mailbox().ReceiveWithTimeout(d, handler)
}

func (ctx *actorContext) Done() <-chan struct{} {
	return ctx.done
}
