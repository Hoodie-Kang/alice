// DKG 종료 전 refresh 를 호출하여 Sign을 위한 하나의 결과를 얻기 위함
package dkg

import (
	"io/ioutil"

	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/cggmp/refresh"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

type refresh_service struct {
	config       *DKGConfig
	refreshInput *dkg.Result
	pm           types.PeerManager

	refresh *refresh.Refresh
	done    chan struct{}
}

func NewRefreshService(config *DKGConfig, refreshInput *dkg.Result, pm types.PeerManager) (*refresh_service, error) {
	s := &refresh_service{
		config:       config,
		refreshInput: refreshInput,
		pm:           pm,
		done:         make(chan struct{}),
	}
	// Create refresh
	// fix me! threshold 2 로 고정해서 진행 -> 차후 *dkg.Result 에 threshold도 넣는 걸로 수정해야함
	refresh, err := refresh.NewRefresh(refreshInput.Share, refreshInput.PublicKey, pm, 2, refreshInput.PartialPubKey, refreshInput.Bks, 2048, refreshInput.SSid, s)
	if err != nil {
		log.Warn("Cannot create a new refresh", "err", err)
		return nil, err
	}
	s.refresh = refresh
	return s, nil
}

func (p *refresh_service) Handle(s network.Stream) {
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

func (p *refresh_service) Process() {
	// 1. Start a refresh process.
	p.refresh.Start()
	defer p.refresh.Stop()
	// 2. Wait the refresh is done or failed
	<-p.done
}

func (p *refresh_service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Refresh failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("Refresh done", "old", oldState.String(), "new", newState.String())
		result, err := p.refresh.GetResult()
		if err == nil {
			writeDKGResult(p.pm.SelfID(), p.config, p.refreshInput, result)
		} else {
			log.Warn("Failed to get result from refresh", "err", err)
		}
		close(p.done)
		return
	}
	// log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
