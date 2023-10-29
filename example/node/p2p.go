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
	"fmt"
	"math/rand"
	"net"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// MakeBasicHost creates a LibP2P host.
func MakeBasicHost(port int64) (host.Host, error) {
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	// sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/172.16.2.117/tcp/%d", port))
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
	// 도메인 - ip 해석
	domainName := "octet.im"

	// 도메인 이름을 IP 주소로 해석
	ipAddresses, err := net.LookupHost(domainName)
	if err != nil {
		fmt.Printf("DNS 조회 오류: %s\n", err)
	}

	// 여러 IP 주소 중 첫 번째 주소 사용 (일반적으로 도메인에는 여러 IP 주소가 연결될 수 있음)
	ipAddress := ipAddresses[0]

	fmt.Printf("도메인 '%s'의 IP 주소: %s\n", domainName, ipAddress)

	return basicHost, nil
}

// getPeerAddr gets peer full address from port.
func GetPeerAddr(port int64, peerId string) string {
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, peerId)
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
	// return fmt.Sprintf("/ip4/172.16.2.117/tcp/%d/p2p/%s", port, pid), nil
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
