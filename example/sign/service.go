// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package sign

import (
	"encoding/hex"
	"io"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/alice/types"
	"github.com/getamis/alice/example/logger"
	"google.golang.org/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

type service struct {
	config *SignConfig
	pm     types.PeerManager

	Sign *sign.Sign
	done chan struct{}
}

func NewService(config *SignConfig, jwt string, pm types.PeerManager) (*service, error) {
	s := &service{
		config: config,
		pm:     pm,
		done:   make(chan struct{}),
	}

	// Inputs from DKG & Refresh Results
	signInput, err := utils.ConvertSignInput(config.Share, config.Pubkey, config.PartialPubKey, config.AllY, config.PaillierKey, config.Ped, config.BKs, config.Message)
	if err != nil {
		logger.Warn("Cannot get SignInput", map[string]string{"err": err.Error()})
		return nil, err
	}
	msg, _ := hex.DecodeString(config.Message)
	// Create sign
	sign, err := sign.NewSign(config.Threshold, config.SSid, signInput.Share, signInput.PublicKey, signInput.PartialPubKey, signInput.Y, signInput.PaillierKey, signInput.PedParameter, signInput.Bks, msg, jwt, pm, s)
	if err != nil {
		logger.Warn("Cannot create a new sign", map[string]string{"err": err.Error()})
		return nil, err
	}
	s.Sign = sign
	return s, nil
}

func (p *service) Handle(s network.Stream) {
	data := &sign.Message{}
	buf, err := io.ReadAll(s)
	if err != nil {
		logger.Warn("Cannot read data from stream", map[string]string{"err": err.Error()})
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, data)
	if err != nil {
		logger.Error("Cannot unmarshal data", map[string]string{"err": err.Error()})
		return
	}

	logger.Info("Received request", map[string]string{"from": s.Conn().RemotePeer().String()})
	err = p.Sign.AddMessage(data.GetId(), data)
	if err != nil {
		logger.Warn("Cannot add message to sign", map[string]string{"err": err.Error()})
		return
	}
}

func (p *service) Process() {
	// 1. Start a sign process.
	p.Sign.Start()
	defer p.Sign.Stop()

	// 2. Wait the sign is done or failed
	<-p.done
}

func (p *service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		logger.Error("Sign failed", map[string]string{"old": oldState.String(), "new": newState.String()})
		close(p.done)
		return
	} else if newState == types.StateDone {
		logger.Info("Sign done", map[string]string{"old": oldState.String(), "new": newState.String()})
		result, err := p.Sign.GetResult()
		if err == nil {
			WriteSignResult(p.pm.SelfID(), result)
		} else {
			logger.Warn("Failed to get result from sign", map[string]string{"err": err.Error()})
		}
		close(p.done)
		return
	}
	logger.Info("State changed", map[string]string{"old": oldState.String(), "new": newState.String()})
}
