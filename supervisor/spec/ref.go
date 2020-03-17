package spec

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
)

type SupRef struct {
	PPID *pid.ProtectedPID
}

func (r *SupRef) CountChildren() (count CountChildren, err error) {
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

func (r *SupRef) DeleteChild(id string) (err error) {
	_, err = r.call(DeleteChild{id})
	return
}

func (r *SupRef) RestartChild(id string) (err error) {
	_, err = r.call(RestartChild{id})
	return
}

func (r *SupRef) StartChild(spec Spec) (err error) {
	_, err = r.call(StartChild{spec})
	return
}

func (r *SupRef) Stop(reason string) (err error) {
	_, err = r.call(Stop{reason})
	return
}

func (r *SupRef) TerminateChild(id string) (err error) {
	_, err = r.call(TerminateChild{id})
	return
}

func (r *SupRef) WithChildren() (childrenInfo WithChildren, err error) {
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

func (r *SupRef) call(request interface{}) (interface{}, error) {
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