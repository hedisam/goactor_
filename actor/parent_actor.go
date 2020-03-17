package actor

// NewParentActor returns an actor with its termination handler that should be called with defer right away
// so the parent actor can handle possible panics and the termination job properly
func NewParentActor() (Actor, func()) {
	actor := createActor()
	return actor, actor.handleTermination
}
