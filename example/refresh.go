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
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/refresh"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

const refreshProtocol = "/refresh/1.0.0"

//export Refresh
func Refresh(argc *C.char, argv *C.char) {
	arg := C.GoString(argv)
	port, _ := strconv.ParseInt(C.GoString(argc), 10, 64)
	config, err := refresh.ReadRefreshConfigFile(arg)
	if err != nil {
		log.Crit("Failed to read config file", "configFile", arg, "err", err)
	}
	config.Port = port
	if config.Peers[0] == 10002 {
		config.Peers[0] = 10004
	} else {
		config.Peers[0] = 10003
	}
	// Make a host that listens on the given multiaddress.
	host, err := peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	// Create a new peer manager.
	pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, refreshProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}

	// Create a new service.
	service, err := refresh.NewService(config, pm)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(refreshProtocol, func(s network.Stream) {
		service.Handle(s)
	})

	// Ensure all peers are connected before starting refresh process.
	pm.EnsureAllConnected()

	// Start refresh process.
	service.Process()

}

func main() {}
