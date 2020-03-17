package ref

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/supervisor/spec"
)

type Ref struct {
	PPID *pid.ProtectedPID
}

func (r *Ref) CountChildren() (CountChildren, error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, Call{
		Sender:  future.Self(),
		Request: CountChildren{},
	})
	result, err := future.Recv()
	if err != nil {return CountChildren{}, err}
	switch result := result.(type) {
	case CountChildren:
		return result, nil
	case Error:
		return CountChildren{}, result
	default:
		return CountChildren{}, errInvalidResponse(result)
	}
}

func (r *Ref) DeleteChild(id string) (err error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, Call{
		Sender:  future.Self(),
		Request: DeleteChild{id},
	})
	result, err := future.Recv()
	if err != nil {return}
	switch result := result.(type) {
	case OK:
		return nil
	case Error:
		return result
	default:
		return errInvalidResponse(result)
	}
}

func (r *Ref) RestartChild(id string) (err error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, Call{
		Sender:  future.Self(),
		Request: RestartChild{id},
	})
	result, err := future.Recv()
	if err != nil {return}
	switch result := result.(type) {
	case OK:
		return nil
	case Error:
		return result
	default:
		return errInvalidResponse(result)
	}
}

func (r *Ref) StartChild(spec spec.Spec) (err error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, Call{
		Sender:  future.Self(),
		Request: StartChild{spec},
	})
	result, err := future.Recv()
	if err != nil {return}
	switch result := result.(type) {
	case OK:
		return nil
	case Error:
		return result
	default:
		return errInvalidResponse(result)
	}
}

func (r *Ref) Stop(reason string) (err error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, Call{
		Sender:  future.Self(),
		Request: Stop{Reason: reason},
	})
	result, err := future.Recv()
	if err != nil {return}
	switch result := result.(type) {
	case OK:
		return nil
	case Error:
		return result
	default:
		return errInvalidResponse(result)
	}
}

func (r *Ref) TerminateChild(id string) (err error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, Call{
		Sender:  future.Self(),
		Request: TerminateChild{id},
	})
	result, err := future.Recv()
	if err != nil {return}
	switch result := result.(type) {
	case OK:
		return nil
	case Error:
		return result
	default:
		return errInvalidResponse(result)
	}
}

func (r *Ref) WithChildren() (WithChildren, error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, WithChildren{})
	result, err := future.Recv()
	if err != nil {return WithChildren{}, err}
	switch result := result.(type) {
	case WithChildren:
		return result, nil
	case Error:
		return WithChildren{}, result
	default:
		return WithChildren{}, errInvalidResponse(result)
	}
}