package vault

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aschmidt75/wireguard-vault-automesh/model"

	log "github.com/sirupsen/logrus"
)

// ReadMeetingPoint accesses vault to read the mesh namework data from the meeting point
func (vc *VaultContext) ReadMeetingPoint(name string) (*model.MeshInfo, error) {
	l := vc.Logical()

	p := DataPath("mp")
	log.WithField("path", p).Trace("Looking for meeting point")

	s, err := l.Read(p)
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token")
		return nil, err
	}
	if s == nil || s.Data["data"] == nil {
		log.Error("No meeting point data found for given network name")
		return nil, nil
	}

	datadata := (s.Data["data"]).(map[string]interface{})
	body := datadata["meshinfo"].(string)

	mi2 := &model.MeshInfo{}
	err = json.Unmarshal([]byte(body), mi2)
	if err != nil {
		log.WithError(err).Error("Error parsing meeting point data")
		return nil, err
	}

	log.WithField("mi", mi2).Debug("meeting point data")

	return mi2, nil
}

// ReadNodes reads the list of nodes from vault
func (vc *VaultContext) ReadNodes() (model.NodeMap, error) {
	l := vc.Logical()

	p := MetaDataPath("nodes")
	log.WithField("path", p).Trace("Looking for nodes...")

	res := make(model.NodeMap, 0)

	s, err := l.List(p)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return res, nil
	}

	keys := s.Data["keys"].([]interface{})

	for _, key := range keys {

		p = DataPath(fmt.Sprintf("nodes/%s", key.(string)))
		v, err := l.Read(p)
		if err != nil {
			return res, err
		}

		if v.Data["data"] == nil {
			log.Error("Internal error, node in vault is present but w/o data.")
		}
		d := v.Data["data"].(map[string]interface{})
		log.WithField("d", d).Trace("ReadNodes.dump")

		lp := 0
		switch d["endpointPort"].(type) {
		case json.Number:
			lp0, err := d["endpointPort"].(json.Number).Int64()
			if err != nil {
				return res, err
			}
			lp = int(lp0)
		case string:
			lp, err = strconv.Atoi(d["endpointPort"].(string))
			if err != nil {
				return res, err
			}
		default:
			log.Error("Unsupported typ")
		}
		res[key.(string)] = model.NodeInfo{
			NodeID:             d["nodeID"].(string),
			WireguardIP:        d["wgip"].(string),
			WireguardPublicKey: d["pubkey"].(string),
			ExternalIP:         d["endpointIP"].(string),
			ListenPort:         int(lp),
		}

	}

	return res, nil
}

// ReadNode reads a single node data from vault
func (vc *VaultContext) ReadNode(key string) (model.NodeInfo, error) {
	l := vc.Logical()

	p := MetaDataPath("nodes")
	log.WithField("path", p).Trace("Looking for nodes...")

	res := model.NodeInfo{}

	p = DataPath(fmt.Sprintf("nodes/%s", key))
	v, err := l.Read(p)
	if err != nil {
		return res, err
	}

	if v.Data["data"] == nil {
		log.Error("Internal error, node in vault is present but w/o data.")
	}
	d := v.Data["data"].(map[string]interface{})
	log.WithField("d", d).Trace("ReadNodes.dump")

	lp, err := d["endpointPort"].(json.Number).Int64()
	if err != nil {
		return res, err
	}
	res = model.NodeInfo{
		NodeID:             d["nodeID"].(string),
		WireguardIP:        d["wgip"].(string),
		WireguardPublicKey: d["pubkey"].(string),
		ExternalIP:         d["endpointIP"].(string),
		ListenPort:         int(lp),
	}

	return res, nil
}
