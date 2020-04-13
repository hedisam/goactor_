package context

import (
	"context"
	"time"
)

type MailboxReceiver interface {
	Receive(handler func(message interface{}) (loop bool))
	ReceiveWithTimeout(d time.Duration, handler func(message interface{}) (loop bool))
}

type Context struct {
	args     []interface{}
	receiver MailboxReceiver
	context  context.Context
}

func NewContext(receiver MailboxReceiver, args []interface{}) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	actorCtx := &Context{
		receiver: receiver,
		args:     args,
		context:  ctx,
	}
	return actorCtx, cancel
}

func (ctx *Context) Args() []interface{} {
	return ctx.args
}

func (ctx *Context) Receive(handler func(message interface{}) (loop bool)) {
	ctx.receiver.Receive(handler)
}

func (ctx *Context) ReceiveWithTimeout(d time.Duration, handler func(message interface{}) (loop bool)) {
	if d < 0 {
		ctx.receiver.Receive(handler)
		return
	}
	ctx.receiver.ReceiveWithTimeout(d, handler)
}

// Done() returns a channel that can be used to know if the actor is been shutdown or not,
// users should listen for the channel in case of long running tasks, if closed, terminate by returning.
func (ctx *Context) Done() <-chan struct{} {
	return ctx.context.Done()
}

// Context returns golang's context.Context object that can be used and passed to inner function calls by the user.
func (ctx *Context) Context() context.Context {
	return ctx.context
}
