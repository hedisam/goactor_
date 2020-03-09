package supervisor

import (
	"fmt"
	"github.com/rs/xid"
)

const (
	// if a child process terminates, only that process is restarted
	OneForOneStrategy Strategy = iota

	// if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	OneForAllStrategy

	// if a child process terminates, the terminated child process and
	// the rest of the spec started after it, are terminated and restarted.
	RestForOneStrategy
)

const (
	defaultMaxRestarts int = 3
	defaultPeriod      int = 5
)

type Strategy int32

type Options struct {
	Strategy    Strategy
	MaxRestarts int
	Period      int
	Name        string
}

func OneForOneStrategyOption() Options {
	return NewOptions(OneForOneStrategy, defaultMaxRestarts, defaultPeriod)
}

func OneForAllStrategyOption() Options {
	return NewOptions(OneForAllStrategy, defaultMaxRestarts, defaultPeriod)
}

func RestForOneStrategyOption() Options {
	return NewOptions(RestForOneStrategy, defaultMaxRestarts, defaultPeriod)
}


func NewOptions(strategy Strategy, maxRestarts, period int) Options {
	return Options{
		Strategy:    strategy,
		MaxRestarts: maxRestarts,
		Period:      period,
		Name:        xid.New().String(),
	}
}

func (opt Options) SetName(name string) Options {
	opt.Name = name
	return opt
}

func (opt *Options) checkOptions() error {
	if opt.Name == "" {
		return fmt.Errorf("invalid supervisor Name: %s", opt.Name)
	} else if opt.Strategy < 0 || opt.Strategy > 2 {
		return fmt.Errorf("invalid Strategy: %d", opt.Strategy)
	} else if opt.Period < 1 {
		return fmt.Errorf("invalid max seconds: %d", opt.Period)
	} else if opt.MaxRestarts < 0 {
		return fmt.Errorf("invalid max restarts: %d", opt.MaxRestarts)
	}

	return nil
}
