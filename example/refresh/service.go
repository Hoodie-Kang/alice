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
package refresh

import (
	"io/ioutil"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

type service struct {
	config *RefreshConfig
	pm     types.PeerManager

	refresh *refresh.Refresh
	done    chan struct{}
}

func NewService(config *RefreshConfig, pm types.PeerManager) (*service, error) {
	s := &service{
		config: config,
		pm:     pm,
		done:   make(chan struct{}),
	}

	// Refresh needs results from DKG.
	dkgResult, err := utils.ConvertDKGResult(config.Pubkey, config.Share, config.BKs, config.PartialPubKey)
	if err != nil {
		log.Warn("Cannot get DKG result", "err", err)
		return nil, err
	}

	// Create refresh
	refresh, err := refresh.NewRefresh(dkgResult.Share, dkgResult.PublicKey, pm, config.Threshold, dkgResult.PartialPubKey, dkgResult.Bks, 2048, config.SSid, s)
	if err != nil {
		log.Warn("Cannot create a new refresh", "err", err)
		return nil, err
	}
	s.refresh = refresh
	return s, nil
}

func (p *service) Handle(s network.Stream) {
	data := &refresh.Message{}
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
	err = p.refresh.AddMessage(data.GetId(), data)
	if err != nil {
		log.Warn("Cannot add message to refresh", "err", err)
		return
	}
}

func (p *service) Process() {
	// 1. Start a refresh process.
	p.refresh.Start()
	defer p.refresh.Stop()

	// 2. Wait the refresh is done or failed
	<-p.done
}

func (p *service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Refresh failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("Refresh done", "old", oldState.String(), "new", newState.String())
		result, err := p.refresh.GetResult()
		if err == nil {
			WriteRefreshResult(p.pm.SelfID(), p.config, result)
		} else {
			log.Warn("Failed to get result from refresh", "err", err)
		}
		close(p.done)
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
