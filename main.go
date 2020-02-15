package main

import (
	"fmt"
	"net"
	"os"
	//"time"

	// "github.com/threefoldtech/zos/pkg/network/dhcp"

	// "github.com/containernetworking/plugins/pkg/utils/sysctl"
	// "github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	// "github.com/threefoldtech/zos/pkg/network/bridge"
	"github.com/threefoldtech/zos/pkg/network/ifaceutil"
	"github.com/threefoldtech/zos/pkg/network/namespace"
	"github.com/vishvananda/netlink"
)

// DefaultBridge2 is the name of the default bridge created by the bootstrap of networkd
const DefaultBridge2 = "zos"

type netconfig struct {
	addresses []*net.IPNet
	routes    []*netlink.Route
}

type found map[string]netconfig

// Bootstrap does a number of things :
//  - get all physical interfaces that have CARRIER
//  - send each interface in a network namespace and do an IP probe

func Bootstrap2() error {
	log.Info().Msg("Starting network discovery")

	links, err := netlink.LinkList()
	if err != nil {
		log.Error().Err(err).Msgf("bootstrap: Couldn't list interfaces")
		return err
	}
	// List through the physical links and bring them up
	for _, link := range ifaceutil.LinkFilter(links, []string{"device"}) {
		device, ok := link.(*netlink.Device)
		if !ok {
			continue
		}
		if device.Name == "lo" || device.Name == "wlp2s0" {
			continue
		}

		log.Info().Msgf("bootstrap: probing interface : %s", device.Name)

		// create an ns, send interface in it
		ifacens, err := namespace.Create(device.Name)
		fmt.Printf("ifacens: %+v\n", ifacens)
		fmt.Printf("ifacens.Fd(): %+v\n", ifacens.Fd())
		if err != nil {
			return err
		}
		defer namespace.Delete(ifacens)

		err = netlink.LinkSetNsFd(device, int(ifacens.Fd()))
		if err != nil {
			log.Error().Err(err).Msgf("bootstrap: Couldn't send %s in namespace : ", device.Name)
		}

	}
	return nil

}

func main() {
	if err := Bootstrap2(); err != nil {
		os.Exit(1)
	}
}
