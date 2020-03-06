package supervisor

import (
	"github.com/hedisam/goactor/internal/pid"
	"time"
)

type registryRepo map[pid.PID]string

type registry struct {
	aliveActors registryRepo
	deadActors  registryRepo
	options 	*options
	// timeTracer contains restart times as unix time
	timeTracer map[string][]int64
}

func newRegistry(ops *options) *registry {
	return &registry{
		aliveActors: make(registryRepo),
		deadActors:  make(registryRepo),
		options:     ops,
		timeTracer:  make(map[string][]int64),
	}
}

// get returns the id associated with a pid. dead is true if the actor has been shutdown by supervisor.
func (r *registry) get(_pid pid.PID) (id string, dead, found bool) {
	id, found = r.aliveActors[_pid]
	if !found {
		id, found = r.deadActors[_pid]
		dead = true
	}
	return
}

// put saves actor's pid and increment restarts count in case of restarts
func (r *registry) put(_pid pid.PID, id string) {
	r.aliveActors[_pid] = id
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
func (r *registry) dead(_pid pid.PID) {
	id, found := r.aliveActors[_pid]
	if !found {
		return
	}
	delete(r.aliveActors, _pid)
	r.deadActors[_pid] = id
}

// reachedMaxRestarts returns true if we have restarts more that allowed in the same period
// notice: the restarts number occurred in the same period is added by one since this method should be called just
// before re-spawning an actor. so we're counting  the not-yet re-spawned one too.
func (r *registry) reachedMaxRestarts(id string) (reached bool) {
	restarts, ok := r.timeTracer[id]
	if !ok {
		// no records yet. the actor has not been started yet.
		return
	}
	// restarts that are not expired, meaning they are in the same period
	var restartsNotEx []int64

	now := time.Now()
	periodStartTime :=  now.Add(time.Duration(-r.options.period) * time.Second).Unix()
	// check how many restarts we've got in the same period
	for _, restartTime := range restarts {
		if restartTime >= periodStartTime {
			// this restart has occurred in the period
			restartsNotEx = append(restartsNotEx, restartTime)
		}
	}
	// added by 1. counting the next restart too that just gonna happen right
	// after returning (if this method return false, of course)
	if len(restartsNotEx) + 1 > r.options.maxRestarts {
		// we got restarts more than the allowed maxRestarts in the same period
		return true
	}

	// get rid of expired timestamps
	r.timeTracer[id] = restartsNotEx
	return
}
