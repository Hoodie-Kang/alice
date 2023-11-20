package node

import (
	"fmt"

	"github.com/getamis/alice/types"
	"github.com/getamis/alice/example/logger"
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
		logger.Error("Protocol failed", map[string]string{"old": oldState.String(), "new": newState.String()})
		return
	} else if newState == types.StateDone {
		l.errCh <- nil
		logger.Info("Protocol done", map[string]string{"old": oldState.String(), "new": newState.String()})
		return
	}
	
	logger.Info("State changed", map[string]string{"old": oldState.String(), "new": newState.String()})
}

func (l *listener) Done() <-chan error {
	return l.errCh
}
