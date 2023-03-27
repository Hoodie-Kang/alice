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
package master

import (
	"io/ioutil"

	"github.com/getamis/alice/crypto/bip32/master"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
)

type service struct {
	config *MasterConfig
	pm     types.PeerManager

	master  *master.Master
	done chan struct{}
}

func NewService(config *MasterConfig, pm types.PeerManager) (*service, error) {
	s := &service{
		config: config,
		pm:     pm,
		done:   make(chan struct{}),
	}
	// fix me! sid need to be different?
	sid := []byte("mastrsid")
	if config.Role == "Alice" {
		m, err := master.NewAlice(pm, sid, config.Rank, circuitPath, s)
		if err != nil {
			log.Warn("Cannot create a new Alice", "config", config, "err", err)
			return nil, err
		}
		s.master = m
	} else if config.Role == "Bob" {
		m, err := master.NewBob(pm, sid, config.Rank, circuitPath, s)
		if err != nil {
			log.Warn("Cannot create a new Bob", "config", config, "err", err)
			return nil, err
		}
		s.master = m
	} else {
		log.Warn("Role must be Alice or Bob", "err", nil)
		return nil, nil
	}	
	return s, nil
}

func (p *service) Handle(s network.Stream) {
	data := &master.Message{}
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
	err = p.master.AddMessage(data.GetId(), data)
	if err != nil {
		log.Warn("Cannot add message to Master", "err", err)
		return
	}
}

func (p *service) Process() {
	// 1. Start a master process.
	p.master.Start()
	defer p.master.Stop()

	// 2. Wait the master is done or failed
	<-p.done
}

func (p *service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("New Master failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("New Master done", "old", oldState.String(), "new", newState.String())
		_, err := p.master.GetResult()
		if err == nil {
			// writeMasterResult(p.config, result)
		} else {
			log.Warn("Failed to get result from Master", "err", err)
		}
		close(p.done)
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
