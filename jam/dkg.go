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
package dkg

import (
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

const dkgProtocol = "/dkg/1.0.0"

func Dkg(arg string) {
	config, err := ReadDKGConfigFile(arg)
	if err != nil {
		log.Crit("Failed to read config file", "configFile", arg, "err", err)
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
	service, err := NewService(config, pm)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
		service.Handle(s)
	})

	// Ensure all peers are connected before starting DKG process.
	pm.EnsureAllConnected()

	// Start DKG process.
	service.Process()

	// For refresh //
	dkgResult, err := service.Dkg.GetResult()
	if err != nil {
		log.Warn("Failed to get result from DKG", "err", err)
	}
	host, err = peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}
	// Create a new peer manager.
	pm2 := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, dkgProtocol)
	err = pm2.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}
	// Create a new service.
	refreshService, err := NewRefreshService(config, dkgResult, pm2)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
		refreshService.Handle(s)
	})
	// Ensure all peers are connected before starting DKG process.
	pm2.EnsureAllConnected()

	refreshService.Process()
}
