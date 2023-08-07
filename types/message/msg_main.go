// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package message

import (
	"context"
	"errors"
	"sync"
	"strconv"

	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/getamis/alice/example/logger"
)

var (
	ErrOldMessage             = errors.New("old message")
	ErrBadMsg                 = errors.New("bad message")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrDupMsg                 = errors.New("duplicate message")
)

type MsgMain struct {
	// logger         log.Logger
	self		   string
	peerNum        uint32
	msgChs         *MsgChans
	state          types.MainState
	currentHandler types.Handler
	listener       types.StateChangedListener

	lock        sync.RWMutex
	handlerLock sync.RWMutex
	cancel      context.CancelFunc
}

func NewMsgMain(id string, peerNum uint32, listener types.StateChangedListener, initHandler types.Handler, msgTypes ...types.MessageType) *MsgMain {
	return &MsgMain{
		self:         id,
		peerNum:        peerNum,
		msgChs:         NewMsgChans(peerNum, msgTypes...),
		state:          types.StateInit,
		currentHandler: initHandler,
		listener:       listener,
	}
}

func (t *MsgMain) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	//nolint:errcheck
	go t.messageLoop(ctx)
	t.cancel = cancel
}

func (t *MsgMain) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.cancel == nil {
		return
	}
	t.cancel()
	t.cancel = nil
}

func (t *MsgMain) AddMessage(senderId string, msg types.Message) error {
	if senderId != msg.GetId() {
		logger.Warn("Different sender", map[string]string{"senderId": senderId, "msgId": msg.GetId()})
		return ErrBadMsg
	}
	currentMsgType := t.GetHandler().MessageType()
	newMessageType := msg.GetMessageType()
	if currentMsgType > newMessageType {
		logger.Warn("Ignore old message", map[string]string{"currentMsgType": strconv.FormatInt(int64(currentMsgType), 10), "newMessageType": strconv.FormatInt(int64(newMessageType), 10)})
		return ErrOldMessage
	}
	return t.msgChs.Push(msg)
}

func (t *MsgMain) GetHandler() types.Handler {
	t.handlerLock.RLock()
	defer t.handlerLock.RUnlock()

	return t.currentHandler
}

func (t *MsgMain) GetState() types.MainState {
	return t.state
}

func (t *MsgMain) messageLoop(ctx context.Context) (err error) {
	defer func() {
		panicErr := recover()

		if err == nil && panicErr == nil {
			_ = t.setState(types.StateDone)
		} else {
			_ = t.setState(types.StateFailed)
		}
		t.Stop()
	}()

	handler := t.GetHandler()
	msgType := handler.MessageType()
	msgCount := uint32(0)
	for {
		// 1. Pop messages
		// 2. Check if the message is handled before
		// 3. Handle the message
		// 4. Check if we collect enough messages
		// 5. If yes, finalize the handler. Otherwise, wait for the next message
		msg, err := t.msgChs.Pop(ctx, msgType)
		if err != nil {
			logger.Warn("Failed to pop message", map[string]string{"err": err.Error()})
			return err
		}
		id := msg.GetId()
		l := log.New("msgType", msgType, "fromId", id)
		if handler.IsHandled(l, id) {
			logger.Warn("The message is handled before", map[string]string{"msgType": strconv.FormatInt(int64(msgType), 10), "fromId": id})
			return ErrDupMsg
		}

		err = handler.HandleMessage(l, msg)
		if err != nil {
			logger.Warn("Failed to save message", map[string]string{"err": err.Error()})
			return err
		}

		msgCount++
		if msgCount < handler.GetRequiredMessageCount() {
			continue
		}

		nextHandler, err := handler.Finalize(l)
		if err != nil {
			logger.Warn("Failed to go to next handler", map[string]string{"err": err.Error()})
			return err
		}
		// if nextHandler is nil, it means we got the final result
		if nextHandler == nil {
			return nil
		}
		t.handlerLock.Lock()
		t.currentHandler = nextHandler
		handler = t.currentHandler
		t.handlerLock.Unlock()
		newType := handler.MessageType()
		logger.Info("Change handler", map[string]string{"oldType": strconv.FormatInt(int64(msgType), 10), "newType": strconv.FormatInt(int64(newType), 10)})
		msgType = newType
		msgCount = uint32(0)
	}
}

func (t *MsgMain) setState(newState types.MainState) error {
	if t.isInFinalState() {
		logger.Warn("Invalid state transition", map[string]string{"old": t.state.String(), "new": newState.String()})
		return ErrInvalidStateTransition
	}

	logger.Info("State changed", map[string]string{"old": t.state.String(), "new": newState.String()})
	oldState := t.state
	t.state = newState
	t.listener.OnStateChanged(oldState, newState)
	return nil
}

func (t *MsgMain) isInFinalState() bool {
	return t.state == types.StateFailed || t.state == types.StateDone
}
