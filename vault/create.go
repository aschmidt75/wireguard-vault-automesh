package vault

import (
	"encoding/json"

	"github.com/aschmidt75/wireguard-vault-automesh/model"

	log "github.com/sirupsen/logrus"
)

// Create accesses vault to create the mesh namework data
func (vc *Context) Create(name string, networkCidr string) (bool, error) {

	mi := model.MeshInfo{
		Name:        name,
		NetworkCIDR: networkCidr,
	}
	log.WithField("meshinfo", mi).Trace("dump")

	l := vc.Logical()

	p := DataPath(name, "mp")
	log.WithField("path", p).Trace("Looking for meeting point")

	s, err := l.Read(p)
	if err != nil {
		log.WithError(err).Error("Error reading from vault. Please check address and token.")
		return false, err
	}
	if s == nil || s.Data["data"] == nil {
		// Not there, create.
		body, err := json.Marshal(mi)
		data := map[string]interface{}{
			"data": map[string]interface{}{
				"meshinfo": string(body),
			},
			"metadata": map[string]interface{}{},
		}
		if err != nil {
			log.WithError(err).Error("Error marshaling data")
			return false, err
		}
		log.WithField("data", data).Trace("writing to vault")
		s, err = l.Write(p, data)
		if err != nil {
			log.WithError(err).Error("Error writing to vault. Please check address and token.")
			return false, err
		}

		return true, nil
	}

	datadata := (s.Data["data"]).(map[string]interface{})
	body := datadata["meshinfo"].(string)

	mi2 := &model.MeshInfo{}
	err = json.Unmarshal([]byte(body), mi2)
	if err != nil {
		log.WithError(err).Error("Error parsing meeting point data")
		return false, err
	}

	log.WithField("mi", mi2).Debug("meeting point data")

	return false, nil
}
