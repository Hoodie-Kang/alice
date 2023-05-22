package main

import (
	"github.com/getamis/alice/example/bip32/master"
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

const masterProtocol = "/master/1.0.0"
const masterRefreshProtocol = "/master-refresh/1.0.0"

func Bip32Master(arg string) {
	config, err := master.ReadMasterConfigFile(arg)
	if err != nil {
		log.Crit("Failed to read config file", "configFile", arg, "err", err)
	}

	host, err := peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, masterProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}

	service, err := master.NewService(config, pm)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	
	host.SetStreamHandler(masterProtocol, func(s network.Stream) {
		service.Handle(s)
	})
	
	pm.EnsureAllConnected()
	service.Process()
	// For refresh //
	masterResult, err := service.Master.GetResult()
	if err != nil {
		log.Warn("Failed to get result from Master", "err", err)
	}
	host, err = peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}
	// Create a new peer manager.
	pm2 := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, masterRefreshProtocol)
	err = pm2.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}
	// Create a new service.
	refreshService, err := master.NewRefreshService(config, masterResult, pm2)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(masterRefreshProtocol, func(s network.Stream) {
		refreshService.Handle(s)
	})
	// Ensure all peers are connected before starting DKG process.
	pm2.EnsureAllConnected()

	refreshService.Process()
}
