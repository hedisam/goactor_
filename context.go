package goactor

import "time"

type MessageHandler func(message interface{}) bool

type TimeoutMessage struct {}

type Context struct {
	pid 	*PID
	fn		ActorFunc
	args	[]interface{}
}

func newContext(pid *PID) *Context {
	return &Context{
		pid:  pid,
	}
}

func (ctx *Context) withActorFunc(fn ActorFunc) *Context {
	ctx.fn = fn
	return ctx
}

func (ctx *Context) withArgs(args []interface{}) *Context {
	ctx.args = args
	return ctx
}

func (ctx *Context) Args() []interface{} {
	return ctx.args
}

func (ctx *Context) Self() *PID {
	return ctx.pid
}

func (ctx *Context) Recv(handler MessageHandler) {
	ctx.pid.mailbox.receive(handler)
}

func (ctx *Context) RecvWithTimeout(d time.Duration, handler MessageHandler) {
	if d < 1 {
		ctx.Recv(handler)
		return
	}
	ctx.pid.mailbox.receiveWithTimeout(d, handler)
}
