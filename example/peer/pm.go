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
package peer

import (
	"context"
	"sync"
	"time"

	"github.com/getamis/alice/example/utils"
	"github.com/getamis/alice/example/logger"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type peerManager struct {
	id       string
	host     host.Host
	protocol protocol.ID
	peers    map[string]string
}

func NewPeerManager(id string, host host.Host, protocol protocol.ID) *peerManager {
	return &peerManager{
		id:       id,
		host:     host,
		protocol: protocol,
		peers:    make(map[string]string),
	}
}

func (p *peerManager) NumPeers() uint32 {
	return uint32(len(p.peers))
}

func (p *peerManager) SelfID() string {
	return p.id
}

func (p *peerManager) PeerIDs() []string {
	ids := make([]string, len(p.peers))
	i := 0
	for id := range p.peers {
		ids[i] = id
		i++
	}
	return ids
}

func (p *peerManager) MustSend(peerID string, message interface{}) {
	err := send(context.Background(), p.host, p.peers[peerID], message, p.protocol)
	if err != nil {
		return
	}
}

// EnsureAllConnected connects the host to specified peer and sends the message to it.
func (p *peerManager) EnsureAllConnected() error {
	var wg sync.WaitGroup
	ch := make(chan error)
	for _, peerAddr := range p.peers {
		wg.Add(1)
		go func () {
			ch <- connectToPeer(p.host, peerAddr, &wg)
		}()
	}
	wg.Wait()
	return <-ch
}

// AddPeers adds peers to peer list.
func (p *peerManager) AddPeers(peerPorts []int64) error {
	for _, peerPort := range peerPorts {
		peerID := utils.GetPeerIDFromPort(peerPort)
		peerAddr, err := getPeerAddr(peerPort)
		if err != nil {
			logger.Warn("Cannot get peer address", map[string]string{"peerID": peerID, "err": err.Error()})
			return err
		}
		p.peers[peerID] = peerAddr
	}
	return nil
}

func connectToPeer(host host.Host, peerAddr string, wg *sync.WaitGroup) error {
	defer wg.Done()
	ch := make(chan bool)

	go func(done chan bool) {
		time.Sleep(60 * time.Second)
		done <- true
	} (ch)

	for {
		// Connect the host to the peer.
		err := connect(context.Background(), host, peerAddr)
		if err != nil {
			select {
			case <-ch:
				return err
			default:
				time.Sleep(3 * time.Second)
				continue
			}
		}
		logger.Info("Successfully connect to peer", map[string]string{"to": peerAddr})
		return nil
	}
}
