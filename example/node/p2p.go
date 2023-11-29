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
	"strconv"
	"strings"
	"encoding/binary"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// MakeBasicHost creates a LibP2P host.
func MakeBasicHost(ip string, port string) (host.Host, error) {
	// sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, port))
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port))
	if err != nil {
		return nil, err
	}
	src := Source(ip, port)
	priv, err := generateIdentity(src)
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

func Source(ip string, port string) int64 {
	var byteArray []byte
	s := strings.Split(ip, ".")
	formattedPort := fmt.Sprintf("%06s", port)

	for _, i := range s {
		intI, _ := strconv.Atoi(i)
		byteArray = append(byteArray, byte(intI))
	}
	byteArray = append(byteArray, 0)
	for i := 0; i < len(formattedPort); i += 2 {
		// 두 글자씩 나눠서 바이트로 변환
		intP, _ := strconv.Atoi(formattedPort[i : i+2])
		bytes := byte(intP)
		// 바이트 배열에 추가
		byteArray = append(byteArray, bytes)
    }
	return int64(binary.BigEndian.Uint64(byteArray))
}

func GetPeerAddr(ip string, port string) (string, error) {
	p := Source(ip, port)
	priv, err := generateIdentity(p)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", ip, port, pid), nil
}

func generateIdentity(src int64) (crypto.PrivKey, error) {
	// Use IP+Port as byte array
	r := rand.New(rand.NewSource(src))

	// Generate a key pair for this host.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		return nil, err
	}
	return priv, nil
}