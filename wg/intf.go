package wg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os/exec"

	log "github.com/sirupsen/logrus"
	wg "golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WireguardInterface is struct
// for all tasks related a given wireguard interface
type WireguardInterface struct {
	InterfaceName string
	IP            net.IP // local ip of wg interface
	EndpointIP    net.IP
	ListenPort    int
	PublicKey     string
}

// HasInterface checks if the interface is already present
func (wgi *WireguardInterface) HasInterface() (bool, error) {
	wgClient, err := wg.New()
	if err != nil {
		return false, err
	}
	defer wgClient.Close()

	device, err := wgClient.Device(wgi.InterfaceName)
	if err != nil {
		return false, err
	}

	return (device != nil), nil
}

// AddInterface adds a new wireguard interface
func (wgi *WireguardInterface) AddInterface() error {
	i, err := net.InterfaceByName(wgi.InterfaceName)

	if i == nil || err != nil {
		// create wireguard interface
		cmd := exec.Command("/sbin/ip", "link", "add", "dev", wgi.InterfaceName, "type", "wireguard")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
		_, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		if len(errStr) > 0 {
			e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
			return errors.New(e)
		}
	}

	i, err = net.InterfaceByName(wgi.InterfaceName)
	if err != nil {
		return err
	}
	log.Tracef("created interface %s", i.Name)

	return nil
}

// EnsureIPAddressIsAssigned checks the local ip of the associated wireguard
// interface. if no ip has been assigned yet, the one from WireguardInterface is assigned.
func (wgi *WireguardInterface) EnsureIPAddressIsAssigned() error {

	i, err := net.InterfaceByName(wgi.InterfaceName)
	if err != nil {
		return err
	}
	log.WithField("intfName", i.Name).Tracef("found wg interface")

	// Assign IP if not yet present
	a, err := i.Addrs()
	if len(a) == 0 {
		cmd := exec.Command("/sbin/ip", "address", "add", "dev", wgi.InterfaceName, wgi.IP.String())
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
		_, errStr := string(stdout.Bytes()), string(stderr.Bytes())
		if len(errStr) > 0 {
			e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
			return errors.New(e)
		}
		log.WithFields(log.Fields{
			"intfName": i.Name,
			"ip":       a[0],
		}).Tracef("added ip to interface")
	}

	a, err = i.Addrs()
	if len(a) == 0 {
		e := fmt.Sprintf("unable to add ip address %s to interface %s: %s", wgi.IP.String(), wgi.InterfaceName, err)
		return errors.New(e)
	}

	return nil
}

func (wgi *WireguardInterface) EnsureInterfaceIsUp() error {

	// bring up wireguard interface
	cmd := exec.Command("/sbin/ip", "link", "set", "up", "dev", wgi.InterfaceName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	_, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}

	return nil
}

func (wgi *WireguardInterface) SetupInterfaceWithConfig() error {
	// wireguard: create private key, add device (listen-port)
	wgClient, err := wg.New()
	if err != nil {
		return err
	}
	defer wgClient.Close()

	wgDevice, err := wgClient.Device(wgi.InterfaceName)
	if err != nil {
		return err
	}

	// check if device already has key and Listen port set. If not, do so
	if bytes.Compare(wgDevice.PrivateKey[:], emptyBytes32) == 0 {
		log.Trace("Private key is empty, generating new key")

		newKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return err
		}

		newConfig := wgtypes.Config{
			PrivateKey: &newKey,
		}
		err = wgClient.ConfigureDevice(wgi.InterfaceName, newConfig)
		if err != nil {
			return err
		}
		log.Trace("Set device private key")
	}

	if wgDevice.ListenPort == 0 {
		newConfig := wgtypes.Config{
			ListenPort: &wgi.ListenPort,
		}
		err = wgClient.ConfigureDevice(wgi.InterfaceName, newConfig)
		if err != nil {
			return err
		}
		log.Trace("Set device listen port")
	}

	// query again make sure stuff is present
	wgDevice, err = wgClient.Device(wgi.InterfaceName)
	if err != nil {
		return err
	}

	if bytes.Compare(wgDevice.PrivateKey[:], emptyBytes32) == 0 || bytes.Compare(wgDevice.PublicKey[:], emptyBytes32) == 0 || wgDevice.ListenPort == 0 {
		return errors.New("unable to set wireguard configuration")
	}

	wgi.PublicKey = base64.StdEncoding.EncodeToString(wgDevice.PublicKey[:])
	log.WithField("pubkey", wgi.PublicKey).Trace("dump")

	return nil
}

// AddPeer adds a new peer to an existing interface
func (wgi *WireguardInterface) AddPeer(remoteEndpointIP string, listenPort int, pubkey string, allowedIPs []net.IPNet, psk *string) error {
	wgClient, err := wg.New()
	if err != nil {
		return err
	}
	defer wgClient.Close()

	var pskAsKey wgtypes.Key
	if psk != nil {
		pskAsKey, err = wgtypes.ParseKey(*psk)
		if err != nil {
			return err
		}
	}

	// process peer
	pk, err := wgtypes.ParseKey(pubkey)
	if err != nil {
		return err
	}
	ep, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteEndpointIP, listenPort))
	if err != nil {
		return err
	}

	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			wgtypes.PeerConfig{
				PublicKey:    pk,
				Remove:       false,
				PresharedKey: &pskAsKey,
				Endpoint:     ep,
				AllowedIPs:   allowedIPs,
			},
		},
	}

	log.WithFields(log.Fields{"new": newConfig}).Trace("adding peer...")
	err = wgClient.ConfigureDevice(wgi.InterfaceName, newConfig)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"intf": wgi.InterfaceName, "PubKey": pubkey}).Info("Added peer.")

	return nil
}

var (
	emptyBytes32 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)
