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
	"io/ioutil"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/sign"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

type service struct {
	config *SignConfig
	pm     types.PeerManager

	sign *sign.Sign
	done chan struct{}
}

func NewService(config *SignConfig, pm types.PeerManager) (*service, error) {
	s := &service{
		config: config,
		pm:     pm,
		done:   make(chan struct{}),
	}

	// Inputs from DKG & Refresh Results
	signInput, err := utils.ConvertSignInput(config.Share, config.Pubkey, config.PartialPubKey, config.AllY, config.PaillierKey, config.Ped, config.BKs, "")
	if err != nil {
		log.Warn("Cannot get SignInput", "err", err)
		return nil, err
	}
	msg, _ := hex.DecodeString(config.Message)
	// Create sign
	sign, err := sign.NewSign(config.Threshold, config.SSid, signInput.Share, signInput.PublicKey, signInput.PartialPubKey, signInput.Y, signInput.PaillierKey, signInput.PedParameter, signInput.Bks, msg, pm, s)
	if err != nil {
		log.Warn("Cannot create a new sign", "err", err)
		return nil, err
	}
	s.sign = sign
	return s, nil
}

func (p *service) Handle(s network.Stream) {
	data := &sign.Message{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		log.Warn("Cannot read data from stream", "err", err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Error("Cannot unmarshal data", "err", err)
		return
	}

	log.Info("Received request", "from", s.Conn().RemotePeer())
	err = p.sign.AddMessage(data.GetId(), data)
	if err != nil {
		log.Warn("Cannot add message to sign", "err", err)
		return
	}
}

func (p *service) Process() {
	// 1. Start a sign process.
	p.sign.Start()
	defer p.sign.Stop()

	// 2. Wait the sign is done or failed
	<-p.done
}

func (p *service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Sign failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("Sign done", "old", oldState.String(), "new", newState.String())
		result, err := p.sign.GetResult()
		if err == nil {
			WriteSignResult(p.pm.SelfID(), result)
		} else {
			log.Warn("Failed to get result from sign", "err", err)
		}
		close(p.done)
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
