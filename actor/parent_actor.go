package actor

// NewParentActor returns an Actor with its termination handler that should be deferred right away
// so the parent Actor can handle possible panics and the termination job properly
func NewParentActor() (*Actor, func()) {
	actor := createActor()
	return actor, actor.handleTermination
}
