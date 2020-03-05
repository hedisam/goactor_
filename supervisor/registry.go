package supervisor

import "github.com/hedisam/goactor/internal/pid"

type registryRepo map[pid.PID]string

type registry struct {
	aliveActors registryRepo
	deadActors  registryRepo
}

func newRegistry() *registry {
	return &registry{
		aliveActors: make(registryRepo),
		deadActors:  make(registryRepo),
	}
}

func (r *registry) get(_pid pid.PID) (id string, dead, found bool) {
	id, found = r.aliveActors[_pid]
	if !found {
		id, found = r.deadActors[_pid]
		dead = true
	}
	return
}

func (r *registry) put(_pid pid.PID, id string) {
	r.aliveActors[_pid] = id
}

func (r *registry) dead(_pid pid.PID) {
	id, found := r.aliveActors[_pid]
	if !found {
		return
	}
	delete(r.aliveActors, _pid)
	r.deadActors[_pid] = id
}
