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
package main

import (
	"C"
	"strconv"
)

import (
	"github.com/getamis/alice/example/dkg"
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

const dkgProtocol = "/dkg/1.0.0"

//export Dkg
func Dkg(argc *C.char, argv *C.char, arg *C.char, info *C.char) int {
	port, _ := strconv.ParseInt(C.GoString(argc), 10, 64)
	peers, _ := strconv.ParseInt(C.GoString(argv), 10, 64)
	path := C.GoString(arg)
	information := C.GoString(info)

	config := &dkg.DKGConfig{
		Port:      port,
		Rank:      0,
		Threshold: 2,
		Peers:     []int64{peers},
	}
	// Make a host that listens on the given multiaddress.
	host, err := peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	// Create a new peer manager.
	pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, dkgProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}

	// Create a new service.
	service, err := dkg.NewService(config, pm, path, information)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
		service.Handle(s)
	})
	// Ensure all peers are connected before starting DKG process.
	err = pm.EnsureAllConnected()
	if err != nil {
		log.Crit("Connection Timeout", "err", err)
		return -1
	}
	// Start DKG process.
	service.Process()
	return 0
}

func main() {}
