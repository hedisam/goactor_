package actor

type exitMessage interface {
	SetRelation(relation string)
}

type connectedActorsRepository map[string]UserPID

type connectedActorsController struct {
	*linkedActors
	*monitorActors
}

func (actors *connectedActorsController) notify(message exitMessage) {
	actors.linkedActors.notify(message)
	actors.monitorActors.notify(message)
}

func newConnectedActorsController() *connectedActorsController {
	return &connectedActorsController{
		linkedActors:  &linkedActors{
			repo: connectedActorsRepository{},
		},
		monitorActors: &monitorActors{
			repo: connectedActorsRepository{},
		},
	}
}


/////////////

type linkedActors struct {
	// actors that are linked to me. two way communication
	repo connectedActorsRepository
}

func (links *linkedActors) link(pid UserPID) {
	links.repo[pid.ID()] = pid
}

func (links *linkedActors) unlink(pid UserPID) {
	delete(links.repo, pid.ID())
}

func (links *linkedActors) notify(message exitMessage) {
	message.SetRelation("linked")
	for _, linked := range links.repo {
		linked.SendSystemMessage(message)
		// we can't shutdown our parent supervisor
		//if shutdown && a.supervisedBy != linked {
		//	linked.Shutdown()
		//}
	}
}


///////////////////////

type monitorActors struct {
	// actors that are monitoring me. one way communication
	repo connectedActorsRepository
}

func (monitors *monitorActors) monitoredBy(pid UserPID) {
	monitors.repo[pid.ID()] = pid
}

func (monitors *monitorActors) demoniteredBy(pid UserPID) {
	delete(monitors.repo, pid.ID())
}

func (monitors *monitorActors) notify(message exitMessage) {
	message.SetRelation("monitored")
	for _, monitor := range monitors.repo {
		monitor.SendSystemMessage(message)
	}
}