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
	"errors"
	"fmt"
	"math/rand"

	"google.golang.org/protobuf/proto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/getamis/alice/example/logger"
)

// MakeBasicHost creates a LibP2P host.
func MakeBasicHost(port int64) (host.Host, error) {
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	// sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err != nil {
		return nil, err
	}

	priv, err := generateIdentity(port)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	return basicHost, nil
}

// getPeerAddr gets peer full address from port.
func getPeerAddr(port int64) (string, error) {
	priv, err := generateIdentity(port)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, pid), nil
}
// Fix me: generateIdentity generates a fixed key pair by using "something(tss 대표지갑 인덱스)" as random source.
func generateIdentity(port int64) (crypto.PrivKey, error) {
	// Use the port as the randomness source in this example.
	// #nosec: G404: Use of weak random number generator (math/rand instead of crypto/rand)
	r := rand.New(rand.NewSource(port))

	// Generate a key pair for this host.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// send sends the proto message to specified peer.
func send(ctx context.Context, host host.Host, target string, data interface{}, protocol protocol.ID) error {
	msg, ok := data.(proto.Message)
	if !ok {
		logger.Error("invalid proto message", map[string]string{})
		return errors.New("invalid proto message")
	}
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

	s, err := host.NewStream(ctx, info.ID, protocol)
	if err != nil {
		logger.Error("Cannot create a new stream", map[string]string{"from": host.ID().String(), "to": target, "err": err.Error()})
		return err
	}

	bs, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("Cannot marshal message", map[string]string{"err": err.Error()})
		return err
	}

	_, err = s.Write(bs)
	if err != nil {
		logger.Warn("Cannot write message to IO", map[string]string{"err": err.Error()})
		return err
	}
	err = s.Close()
	if err != nil {
		logger.Warn("Cannot close the stream", map[string]string{"err": err.Error()})
		return err
	}

	logger.Info("Sent message", map[string]string{"to": target})
	return nil
}

// connect connects the host to the specified peer.
func connect(ctx context.Context, host host.Host, target string) error {
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
		logger.Warn("Failed to connect to peer", map[string]string{"err": err.Error()})
		return err
	}
	return nil
}
