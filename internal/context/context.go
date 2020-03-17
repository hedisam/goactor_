package context

import (
	"context"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"time"
)

type Context struct {
	pid  pid.PID
	args []interface{}
	// use context.Context instead of done channel
	ctx context.Context
}

func NewContext(pid pid.PID, args []interface{}) *Context {
	ctx, cancel := context.WithCancel(context.Background())
	actorCtx := &Context{
		pid:  pid,
		args: args,
		ctx: ctx,
	}
	pid.SetShutdownFn(cancel)
	return actorCtx
}

func (ctx *Context) Args() []interface{} {
	return ctx.args
}

func (ctx *Context) Receive(handler mailbox.MessageHandler) {
	ctx.pid.Mailbox().Receive(handler)
}

func (ctx *Context) ReceiveWithTimeout(d time.Duration, handler mailbox.MessageHandler) {
	if d < 1 {
		ctx.pid.Mailbox().Receive(handler)
		return
	}
	ctx.pid.Mailbox().ReceiveWithTimeout(d, handler)
}

// Done() returns a channel that can be used to know if the actor is been shutdown or not,
// users should listen for the channel in case of long running tasks, if closed, terminate by returning.
func (ctx *Context) Done() <-chan struct{} {
	return ctx.ctx.Done()
}

// Context returns golang's context.Context that can be used and passed to inner function calls by the user.
func (ctx *Context) Context() context.Context {
	return ctx.ctx
}
