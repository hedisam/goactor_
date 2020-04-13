package supervisor

import (
	"github.com/hedisam/goactor/supervisor/spec"
	"time"
)

type registryRepo map[spec.CancelablePID]string

type registry struct {
	aliveActors registryRepo
	deadActors  registryRepo
	options 	*Options
	// timeTracer contains restart times as unix time
	timeTracer map[string][]int64
}

func newRegistry(ops *Options) *registry {
	return &registry{
		aliveActors: make(registryRepo),
		deadActors:  make(registryRepo),
		options:     ops,
		timeTracer:  make(map[string][]int64),
	}
}

// id returns the id associated with a pid. dead is true if the actor has been shutdown by supervisor.
func (r *registry) id(pid spec.CancelablePID) (id string, dead, found bool) {
	id, found = r.aliveActors[pid]
	if !found {
		id, found = r.deadActors[pid]
		dead = true
	}
	return
}

// alivePID returns the pid.PID associated with the id if the actor is alive
func (r *registry) alivePID(id string) (spec.CancelablePID, bool) {
	for pid, _id := range r.aliveActors {
		if _id == id {
			return pid, true
		}
	}
	return nil, false
}

// put saves actor's pid and increment restarts count in case of restarts
func (r *registry) put(pid spec.CancelablePID, id string) {
	r.aliveActors[pid] = id
	restarts, ok := r.timeTracer[id]
	if !ok {
		// it's the first time this actor been spawned. so it's not a restart. we only record restarts timestamp
		r.timeTracer[id] = []int64{}
		return
	}
	// append the current restart's timestamp
	r.timeTracer[id] = append(restarts, time.Now().Unix())
}

// dead declares an actor dead by its pid
func (r *registry) dead(pid spec.CancelablePID) {
	id, found := r.aliveActors[pid]
	if !found {
		return
	}
	delete(r.aliveActors, pid)
	r.deadActors[pid] = id
}

// reachedMaxRestarts returns true if we have restarts more than allowed in the same Period
// notice: the restarts number occurred in the same Period is added by one since this method should be called just
// before re-spawning an actor. so we're counting  the not-yet re-spawned one too.
func (r *registry) reachedMaxRestarts(id string) (reached bool) {
	restarts, ok := r.timeTracer[id]
	if !ok {
		// no records yet. the actor has not been started yet.
		return
	}
	// restarts that are not expired, meaning they are in the same Period
	var restartsNotEx []int64

	now := time.Now()
	periodStartTime :=  now.Add(time.Duration(-r.options.Period) * time.Second).Unix()
	// check how many restarts we've got in the same Period
	for _, restartTime := range restarts {
		if restartTime >= periodStartTime {
			// this restart has occurred in the Period
			restartsNotEx = append(restartsNotEx, restartTime)
		}
	}
	// added by 1. counting the next restart too that just gonna happen right
	// after returning (if this method return false, of course)
	if len(restartsNotEx) + 1 > r.options.MaxRestarts {
		// we got restarts more than the allowed MaxRestarts in the same Period
		return true
	}

	// id rid of expired timestamps
	r.timeTracer[id] = restartsNotEx
	return
}
