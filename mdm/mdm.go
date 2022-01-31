package mdm

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/groob/plist"
	macospkg "github.com/korylprince/go-macos-pkg"
	"github.com/korylprince/ls-relay-cert/profile"
	"go.mozilla.org/pkcs7"
	"golang.org/x/crypto/pkcs12"
)

var ErrNotFound = errors.New("serial not found")

type Config struct {
	MDMPrefix       string
	MDMToken        string
	SigningIdentity string
	CacheSize       int
	CacheTTL        time.Duration
	CachePrefix     string
	*profile.Config
}

// MDM is a MicroMDM service
type MDM struct {
	*Config
	cert *x509.Certificate
	key  *rsa.PrivateKey
	*FileStore
}

func New(config *Config) (*MDM, error) {
	// read "Apple Developer ID Installer" identity
	identity, err := os.ReadFile(config.SigningIdentity)
	if err != nil {
		return nil, fmt.Errorf("could not read identity: %w", err)
	}
	key, cert, err := pkcs12.Decode(identity, "")
	if err != nil {
		return nil, fmt.Errorf("could not decode identity: %w", err)
	}

	return &MDM{
		Config:    config,
		cert:      cert,
		key:       key.(*rsa.PrivateKey),
		FileStore: NewFileStore(config.CacheSize, config.CacheTTL),
	}, nil
}

// SerialToUDID returns the UDID for the given serial. If the serial is not found, ErrNotFound is returned
func (m *MDM) SerialToUDID(serial string) (string, error) {
	type response struct {
		Devices []struct {
			UDID string `json:"udid"`
		} `json:"devices"`
		Error string `json:"error"`
	}

	q := map[string]interface{}{
		"filter_serial": []string{serial},
	}

	j, err := json.Marshal(q)
	if err != nil {
		return "", fmt.Errorf("could not marshal query: %w", err)
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/devices", m.MDMPrefix), bytes.NewBuffer(j))
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	r.SetBasicAuth("micromdm", m.MDMToken)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", fmt.Errorf("could not complete request: %w", err)
	}
	defer res.Body.Close()

	resp := new(response)
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(resp); err != nil {
		return "", fmt.Errorf("could not parse response: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("could not query devices: %s", resp.Error)
	}

	if len(resp.Devices) != 1 || resp.Devices[0].UDID == "" {
		return "", ErrNotFound
	}

	return resp.Devices[0].UDID, nil
}

// Command runs the MDM cmd
func (m *MDM) Command(cmd map[string]interface{}) error {
	type response struct {
		Error string `json:"error"`
	}

	j, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("could not marshal command: %w", err)
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/commands", m.MDMPrefix), bytes.NewBuffer(j))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	r.SetBasicAuth("micromdm", m.MDMToken)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("could not complete request: %w", err)
	}
	defer res.Body.Close()

	resp := new(response)
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(resp); err != nil {
		return fmt.Errorf("could not parse response: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("could not execute command: %s", resp.Error)
	}

	return nil
}

// InstallProfile runs the InstallProfile command with the given udid and profile
func (m *MDM) InstallProfile(udid string, profile *profile.TopLevelProfile) error {
	// marshal plist
	buf, err := plist.Marshal(profile)
	if err != nil {
		return fmt.Errorf("could not marshal plist: %w", err)
	}

	// sign plist
	sd, err := pkcs7.NewSignedData(buf)
	if err != nil {
		return fmt.Errorf("could not init pkcs7: %w", err)
	}

	if err := sd.AddSigner(m.cert, m.key, pkcs7.SignerInfoConfig{}); err != nil {
		return fmt.Errorf("could not add pkcs7 signer: %w", err)
	}

	signed, err := sd.Finish()
	if err != nil {
		return fmt.Errorf("could not marshal pkcs7: %w", err)
	}

	// execute command
	cmd := map[string]interface{}{
		"request_type": "InstallProfile",
		"udid":         udid,
		"payload":      signed,
	}

	if err = m.Command(cmd); err != nil {
		return fmt.Errorf("could not execute InstallProfile command: %w", err)
	}

	return nil
}

// InstallEnterpriseApplication runs the InstallEnterpriseApplication command with the given udid and manifest
func (m *MDM) InstallEnterpriseApplication(udid string, manifest *macospkg.Manifest) error {
	cmd := map[string]interface{}{
		"request_type": "InstallEnterpriseApplication",
		"udid":         udid,
		"manifest":     manifest,
	}

	if err := m.Command(cmd); err != nil {
		return fmt.Errorf("could not execute InstallEnterpriseApplication command: %w", err)
	}

	return nil
}
