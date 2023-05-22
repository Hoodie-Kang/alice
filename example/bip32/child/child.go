package child

import (
	"github.com/getamis/alice/example/peer"
	"github.com/getamis/alice/example/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
)

const childProtocol = "/child/1.0.0"

func Bip32Child(arg string) {
	config, err := ReadChildConfigFile(arg)
	if err != nil {
		log.Crit("Failed to read config file", "configFile", arg, "err", err)
	}

	host, err := peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, childProtocol)
	err = pm.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}

	service, err := NewService(config, pm)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	
	host.SetStreamHandler(childProtocol, func(s network.Stream) {
		service.Handle(s)
	})
	
	pm.EnsureAllConnected()
	service.Process()	
	// For refresh //
	masterResult, err := service.Child.GetResult()
	if err != nil {
		log.Warn("Failed to get result from Master", "err", err)
	}
	host, err = peer.MakeBasicHost(config.Port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}
	// Create a new peer manager.
	pm2 := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, childProtocol)
	err = pm2.AddPeers(config.Peers)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}
	// Create a new service.
	refreshService, err := NewRefreshService(config, masterResult, pm2)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(childProtocol, func(s network.Stream) {
		refreshService.Handle(s)
	})
	// Ensure all peers are connected before starting DKG process.
	pm2.EnsureAllConnected()

	refreshService.Process()
}