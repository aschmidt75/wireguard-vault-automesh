package wg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"

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

	var err error

	i, err := net.InterfaceByName(wgi.InterfaceName)
	if err != nil {
		return err
	}
	log.WithField("intfName", i.Name).Tracef("found wg interface")

	// Assign IP if not yet present
	a, err := i.Addrs()
	if err != nil {
		return err
	}
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
	}

	a, err = i.Addrs()
	if len(a) == 0 {
		e := fmt.Sprintf("unable to add ip address %s to interface %s: %s", wgi.IP.String(), wgi.InterfaceName, err)
		return errors.New(e)
	}
	log.WithFields(log.Fields{
		"intfName": i.Name,
		"ip":       a[0],
	}).Tracef("added ip to interface")

	return nil
}

// EnsureInterfaceIsUp checks if the wireguard interface is up. if not, up's it. all using /sbin/ip
func (wgi *WireguardInterface) EnsureInterfaceIsUp() error {

	// bring up wireguard interface
	cmd := exec.Command("/sbin/ip", "--br", "link", "show", "dev", wgi.InterfaceName, "up", "type", "wireguard")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}
	if len(outStr) > 0 {
		// output should contain interface name an "UP" TODO
		log.WithField("o", outStr).Trace("Interface is up")
		return nil
	}

	// bring up wireguard interface
	cmd = exec.Command("/sbin/ip", "link", "set", "up", "dev", wgi.InterfaceName)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	_, errStr = string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}

	return nil
}

// EnsureRouteIsSet checks if there is a route to given network. If not, adds it. all using /sbin/ip
func (wgi *WireguardInterface) EnsureRouteIsSet(networkCIDR string) error {

	//
	cmd := exec.Command("/sbin/ip", "route", "show", "dev", wgi.InterfaceName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}
	a := strings.Split(outStr, " ")
	if len(a) > 0 {
		if a[0] == networkCIDR {
			log.WithField("o", outStr).Trace("Route present")
			return nil
		}
	}

	//
	cmd = exec.Command("/sbin/ip", "route", "add", networkCIDR, "dev", wgi.InterfaceName)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	_, errStr = string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}

	return nil
}

// SetupInterfaceWithConfig makes sure that the wireguard interface
// has a keypair and a listen port configured. Extracts public key
// part and stores it in wgi.
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
		if wgi.ListenPort == 0 {
			return errors.New("wg listenPort may not be 0")
		}

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
	if wgDevice == nil {
		return errors.New("error reading wg device configuration")
	}
	//log.WithField("wgd", *wgDevice).Trace("SetupInterfaceWithConfig.dump") // caution: dumps private key

	if bytes.Compare(wgDevice.PrivateKey[:], emptyBytes32) == 0 || bytes.Compare(wgDevice.PublicKey[:], emptyBytes32) == 0 || wgDevice.ListenPort == 0 {
		return errors.New("unable to set wireguard key configuration")
	}

	wgi.PublicKey = base64.StdEncoding.EncodeToString(wgDevice.PublicKey[:])
	log.WithField("pubkey", wgi.PublicKey).Trace("SetupInterfaceWithConfig.dump")

	return nil
}

// AddPeer adds a new peer to an existing interface
func (wgi *WireguardInterface) AddPeer(remoteEndpointIP string, listenPort int, pubkey string, allowedIPs []net.IPNet, psk *string) (bool, error) {
	wgClient, err := wg.New()
	if err != nil {
		return false, err
	}
	defer wgClient.Close()

	pk, err := wgtypes.ParseKey(pubkey)
	if err != nil {
		return false, err
	}

	wgDevice, err := wgClient.Device(wgi.InterfaceName)
	if err != nil {
		return false, err
	}
	for _, peer := range wgDevice.Peers {
		if peer.PublicKey == pk {
			log.WithField("pubkey", pubkey).Trace("Already present, skipping")
			return false, nil
		}
	}

	var pskAsKey wgtypes.Key
	if psk != nil {
		pskAsKey, err = wgtypes.ParseKey(*psk)
		if err != nil {
			return false, err
		}
	}

	// process peer
	ep, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", remoteEndpointIP, listenPort))
	if err != nil {
		return false, err
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
		return false, err
	}
	log.WithFields(log.Fields{"intf": wgi.InterfaceName, "PubKey": pubkey}).Info("Added peer.")

	return true, nil
}

// RemoveWgInterface takes down an existing wireguard interface
func (wgi *WireguardInterface) RemoveWgInterface() error {
	i, err := net.InterfaceByName(wgi.InterfaceName)

	if err != nil {
		return err
	}
	if i == nil {
		e := fmt.Sprintf("No network/interface by name %s", wgi.InterfaceName)
		return errors.New(e)
	}

	// remove all peers (necessary?)

	// take down wireguard interface
	cmd := exec.Command("/sbin/ip", "link", "set", "down", "dev", wgi.InterfaceName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	_, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}

	// remove wireguard interface
	cmd = exec.Command("/sbin/ip", "link", "delete", "dev", wgi.InterfaceName, "type", "wireguard")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	_, errStr = string(stdout.Bytes()), string(stderr.Bytes())
	if len(errStr) > 0 {
		e := fmt.Sprintf("/sbin/ip reported: %s", errStr)
		return errors.New(e)
	}

	log.WithFields(log.Fields{"intf": wgi.InterfaceName}).Info("Removed wireguard interface")

	return nil
}

// RemoveAllWgPeers takes down all peers of an existing interface
func (wgi *WireguardInterface) RemoveAllWgPeers() error {
	wgClient, err := wg.New()
	if err != nil {
		return err
	}
	defer wgClient.Close()

	newConfig := wgtypes.Config{
		ReplacePeers: true,
		Peers:        []wgtypes.PeerConfig{},
	}

	log.WithFields(log.Fields{"new": newConfig}).Trace("removing all peers...")
	err = wgClient.ConfigureDevice(wgi.InterfaceName, newConfig)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"intf": wgi.InterfaceName}).Info("Removed all peers.")

	return nil
}

// RemoveWgPeer takes down a single peer of an existing interface
func (wgi *WireguardInterface) RemoveWgPeer(pubkey string) error {
	wgClient, err := wg.New()
	if err != nil {
		return err
	}
	defer wgClient.Close()

	// process peer
	pk, err := wgtypes.ParseKey(pubkey)
	if err != nil {
		return err
	}

	newConfig := wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			wgtypes.PeerConfig{
				PublicKey: pk,
				Remove:    true,
			},
		},
	}

	log.WithFields(log.Fields{"new": newConfig}).Trace("removing peer...")
	err = wgClient.ConfigureDevice(wgi.InterfaceName, newConfig)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"intf": wgi.InterfaceName, "PubKey": pubkey}).Info("Removed peer.")

	return nil
}

// IterateWgPeerFunc is a callback
type IterateWgPeerFunc func(pubkey string)

// IterateWgPeers reads all peers from given device and calls the func
func (wgi *WireguardInterface) IterateWgPeers(cb IterateWgPeerFunc) error {
	wgClient, err := wg.New()
	if err != nil {
		return err
	}
	defer wgClient.Close()

	wgDevice, err := wgClient.Device(wgi.InterfaceName)
	if err != nil {
		return err
	}
	for _, peer := range wgDevice.Peers {
		cb(base64.StdEncoding.EncodeToString(peer.PublicKey[:]))
	}

	return nil
}

var (
	emptyBytes32 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)
