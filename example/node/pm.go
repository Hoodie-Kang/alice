// Copyright © 2023 AMIS Technologies
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

package node

import (
	"context"
	"sync"
	"time"

	"github.com/getamis/alice/example/logger"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
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
	// sort.Strings(ids)
	return ids
}

func (p *peerManager) MustSend(peerId string, message interface{}) {
	msg, ok := message.(proto.Message)
	if !ok {
		logger.Error("invalid proto message", map[string]string{})
	}

	target := p.peers[peerId]

	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(target)
	if err != nil {
		logger.Error("Cannot parse the target address", map[string]string{"target": target, "err": err.Error()})
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		logger.Error("Cannot parse addr", map[string]string{"addr": maddr.String(), "err": err.Error()})
	}

	s, err := p.host.NewStream(context.Background(), info.ID, p.protocol)
	if err != nil {
		logger.Error("Cannot create a new stream", map[string]string{"from": p.host.ID().String(), "to": info.ID.String(), "protocol": string(p.protocol), "err": err.Error()})
	}

	bs, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("Cannot marshal message", map[string]string{"err": err.Error()})
	}

	_, err = s.Write(bs)
	if err != nil {
		logger.Error("Cannot write message to IO", map[string]string{"err": err.Error()})
	}

	err = s.Close()
	if err != nil {
		logger.Error("Cannot close the stream", map[string]string{"err": err.Error()})
	}

	logger.Info("Sent message", map[string]string{"peer": target})
}

// EnsureAllConnected connects the host to specified peer and sends the message to it.
func (p *peerManager) EnsureAllConnected() {
	var wg sync.WaitGroup

	// connect connects the host to the specified peer.
	connect := func(ctx context.Context, host host.Host, target string) error {
		// Turn the destination into a multiaddr.
		maddr, err := multiaddr.NewMultiaddr(target)
		if err != nil {
			logger.Error("Cannot parse the target address", map[string]string{"target": target, "err": err.Error()})
			return err
		}

		// Extract the peer ID from the multiaddr.
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			logger.Error("Cannot parse addr", map[string]string{"addr": maddr.String(), "err": err.Error()})
			return err
		}

		// Connect the host to the peer.
		err = host.Connect(ctx, *info)
		if err != nil {
			// log.Warn("Failed to connect to peer", "err", err)
			return err
		}

		for {
			protocols, err := host.Peerstore().GetProtocols(info.ID)
			if err != nil {
				logger.Warn("Failed to get peer protocols", map[string]string{"err": err.Error()})
				return err
			}

			for _, prootocol := range protocols {
				if prootocol == p.protocol {
					return nil
				}
			}

			logger.Info("Waiting for peers", map[string]string{})
			time.Sleep(2 * time.Second)
		}
	}

	for _, peerAddr := range p.peers {
		wg.Add(1)
		addr := peerAddr

		go func() {
			defer wg.Done()
			deadline := time.Now().Add(5 * time.Second)
			for {
				// Connect the host to the peer. with timeout
				ctx, cancel := context.WithDeadline(context.Background(), deadline)
				defer cancel()
				if ctx.Err() == context.DeadlineExceeded {
					logger.Error("Connection Timeout", map[string]string{"err": ctx.Err().Error(), "addr": addr})
				}
				err := connect(ctx, p.host, addr)
				if err != nil {
					time.Sleep(3 * time.Second)
					continue
				}
				logger.Info("Successfully connect to peer", map[string]string{"peer": addr})
				return
			}
		}()
	}
	wg.Wait()
}

// AddPeer adds a peer to the peer list.
func (p *peerManager) AddPeer(serverIP string, serverPort string) {
	peerAddr, _ := GetPeerAddr(serverIP, serverPort)
	p.peers["Octet"] = peerAddr
}
