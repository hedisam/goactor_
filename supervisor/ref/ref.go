package ref

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/supervisor/spec"
)

type Ref struct {
	PPID *pid.ProtectedPID
}

func (r *Ref) CountChildren() (count CountChildren, err error) {
	result, err := r.call(CountChildren{})
	if err != nil {
		return
	}
	count, ok := result.(CountChildren)
	if !ok {
		return count, errInvalidResponse(result)
	}
	return
}

func (r *Ref) DeleteChild(id string) (err error) {
	_, err = r.call(DeleteChild{id})
	return
}

func (r *Ref) RestartChild(id string) (err error) {
	_, err = r.call(RestartChild{id})
	return
}

func (r *Ref) StartChild(spec spec.Spec) (err error) {
	_, err = r.call(StartChild{spec})
	return
}

func (r *Ref) Stop(reason string) (err error) {
	_, err = r.call(Stop{reason})
	return
}

func (r *Ref) TerminateChild(id string) (err error) {
	_, err = r.call(TerminateChild{id})
	return
}

func (r *Ref) WithChildren() (childrenInfo WithChildren, err error) {
	result, err := r.call(WithChildren{})
	if err != nil {
		return
	}
	childrenInfo, ok := result.(WithChildren)
	if !ok {
		return childrenInfo, errInvalidResponse(result)
	}
	return
}

func (r *Ref) call(request interface{}) (interface{}, error) {
	future := actor.NewFutureActor()
	future.Send(r.PPID, request)
	result, err := future.Recv()
	if err != nil {
		return nil, err
	}

	switch result := result.(type) {
	case OK:
		return nil, nil
	case Error:
		return nil, result
	default:
		// call specific response
		return result, nil
	}
}