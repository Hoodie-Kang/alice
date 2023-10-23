package node

import (
	"fmt"

	"github.com/getamis/alice/types"
	logger "github.com/getamis/sirius/log"
)

type Listener interface {
	types.StateChangedListener

	Done() <-chan error
}

func NewListener() *listener {
	return &listener{
		errCh: make(chan error, 1),
	}
}

type listener struct {
	errCh chan error
}

func (l *listener) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		l.errCh <- fmt.Errorf("state %s -> %s", oldState.String(), newState.String())
		return
	} else if newState == types.StateDone {
		l.errCh <- nil
		logger.Info("Protocol done", "oldState", oldState, "newState", newState)
		return
	}
	
	logger.Debug("State changed", "oldState", oldState, "newState", newState)
}

func (l *listener) Done() <-chan error {
	return l.errCh
}
