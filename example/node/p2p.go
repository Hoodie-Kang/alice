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
	"crypto/sha256"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/crypto/hkdf"
)

// MakeBasicHost creates a LibP2P host.
func MakeBasicHost(port int64) (host.Host, error) {
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err != nil {
		return nil, err
	}

	src := Source("127.0.0.1", strconv.Itoa(int(port)))
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

func Source(ip string, port string) []byte {
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
	return byteArray
}

// getPeerAddr gets peer full address from port.
func GetPeerAddr(port int64) (string, error) {
	p := Source("127.0.0.1", strconv.Itoa(int(port)))
	priv, err := generateIdentity(p)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, pid), nil
}

func generateIdentity(src []byte) (crypto.PrivKey, error) {
	// Use IP+Port as byte array
	deterministicHKDFReader := newDeterministicReader(src)
	// Generate a key pair for this host.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, deterministicHKDFReader)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

type deterministicReader struct {
	reader           io.Reader
	singleByteReader io.Reader
}

func newDeterministicReader(seed []byte) io.Reader {
	reader := hkdf.New(sha256.New, seed, nil, nil)
	singleByteReader := hkdf.New(sha256.New, seed, nil, nil)

	return &deterministicReader{
		reader:           reader,
		singleByteReader: singleByteReader,
	}
}

func (r *deterministicReader) Read(p []byte) (n int, err error) {
	if len(p) == 1 {
		return r.singleByteReader.Read(p)
	}
	return r.reader.Read(p)
}