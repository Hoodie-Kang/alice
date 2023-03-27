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
package child

import (
	"io/ioutil"

	"github.com/getamis/alice/crypto/bip32/child"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
)

type service struct {
	config *ChildConfig
	pm     types.PeerManager

	child  *child.Child
	done chan struct{}
}

func NewService(config *ChildConfig, pm types.PeerManager) (*service, error) {
	s := &service{
		config: config,
		pm:     pm,
		done:   make(chan struct{}),
	}

	// Child need results from Master.
	masterResult, err := utils.ConvertMasterResult(config.Role, config.Pubkey, config.Share, config.BKs, config.Seed, config.ChainCode)
	if err != nil {
		log.Warn("Cannot get Master result", "err", err)
		return nil, err
	}
	// fix me! childIndex ? clarify the meaning and the usage
	// _ childIndex 값에 따라서 pubkey, share 등이 결정되는 듯.
	// 2147583650 으로 여러번 시행해도 같은 결과가 나옴.
	// childIndex로 child 를 복구하는 것이 가능할지도?
	// CGGMP에선 refresh 거쳐야해서 복구가 쉽지 않음?
	childIndex := uint32(2147483650)
	// fix me! sid need to be different?
	sid := []byte("childsid")
	if config.Role == "Alice" {
		c, err := child.NewAlice(pm, sid, masterResult.Share, masterResult.Bks, circuitPath, masterResult.ChainCode, config.Depth, childIndex, masterResult.PublicKey, s)
		if err != nil {
			log.Warn("Cannot create a new child - Alice", "config", config, "err", err)
			return nil, err
		}
		s.child = c
	} else if config.Role == "Bob" {
		c, err := child.NewBob(pm, sid, masterResult.Share, masterResult.Bks, circuitPath, masterResult.ChainCode, config.Depth, childIndex, masterResult.PublicKey, s)
		if err != nil {
			log.Warn("Cannot create a new child - Bob", "config", config, "err", err)
			return nil, err
		}
		s.child = c
	} else {
		log.Warn("Role must be Alice or Bob", "err", nil)
		return nil, nil
	}	
	return s, nil
}

func (p *service) Handle(s network.Stream) {
	data := &child.Message{}
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
	err = p.child.AddMessage(data.GetId(), data)
	if err != nil {
		log.Warn("Cannot add message to Child", "err", err)
		return
	}
}

func (p *service) Process() {
	// 1. Start a child process.
	p.child.Start()
	defer p.child.Stop()

	// 2. Wait the child is done or failed
	<-p.done
}

func (p *service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("New Child failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("New Child done", "old", oldState.String(), "new", newState.String())
		_, err := p.child.GetResult()
		if err != nil {
			log.Warn("Failed to get result from Child", "err", err)
		}
		close(p.done)
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
