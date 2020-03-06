package supervisor

import (
	"github.com/rs/xid"
)

const (
	// if a child process terminates, only that process is restarted
	OneForOneStrategy Strategy = iota

	// if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	OneForAllStrategy

	// if a child process terminates, the terminated child process and
	// the rest of the children started after it, are terminated and restarted.
	RestForOneStrategy
)

const (
	defaultMaxRestarts int = 3
	defaultPeriod      int = 5
)

type Strategy int32

type options struct {
	strategy    Strategy
	maxRestarts int
	period      int
	name        string
}

var OneForOneStrategyOption 	= NewOptions(OneForOneStrategy, defaultMaxRestarts, defaultPeriod)
var OneForAllStrategyOption 	= NewOptions(OneForAllStrategy, defaultMaxRestarts, defaultPeriod)
var RestForOneStrategyOption	= NewOptions(RestForOneStrategy, defaultMaxRestarts, defaultPeriod)

func NewOptions(strategy Strategy, maxRestarts, period int) *options {
	return &options{
		strategy:    strategy,
		maxRestarts: maxRestarts,
		period:      period,
		name:        xid.New().String(),
	}
}

func (o *options) SetName(name string) *options {
	o.name = name
	return o
}
