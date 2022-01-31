package mdm

import (
	"bytes"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	"text/template"

	macospkg "github.com/korylprince/go-macos-pkg"
	"github.com/korylprince/ls-relay-cert/cert"
	"github.com/korylprince/ls-relay-cert/profile"
)

//go:embed payload.sh
var payloadScript string
var tmplPostinstall = template.Must(template.New("payload.sh").Parse(payloadScript))

// GeneratePKI generates and returns a certificate root profile and PEM encoded CA and localhost key pairs with the given CA validity in years
func GeneratePKI(years int, config *profile.Config) (*profile.TopLevelProfile, *Payload, error) {
	c, ck, err := cert.GenerateCA(years)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate CA key pair: %w", err)
	}

	lh, lhk, err := cert.GenerateLocalhost(c, ck)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate localhost key pair: %w", err)
	}

	profile, err := profile.New(config, c)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate profile: %w", err)
	}

	return profile, &Payload{
		CA:           string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})),
		CAKey:        string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(ck)})),
		Localhost:    string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: lh.Raw})),
		LocalhostKey: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(lhk)})),
	}, nil
}

// Payload is the script payload containing CA and localhost key pairs
type Payload struct {
	CA           string
	CAKey        string
	Localhost    string
	LocalhostKey string
}

// Deliver generates the necessary profile and certificates and delivers them to the device with serial
func (m *MDM) Deliver(serial string) error {
	udid, err := m.SerialToUDID(serial)
	if err != nil {
		return fmt.Errorf("could not get UDID: %w", err)
	}

	profile, payload, err := GeneratePKI(10, m.Config.Config)
	if err != nil {
		return fmt.Errorf("could not generate pki: %w", err)
	}

	postinstall := new(bytes.Buffer)
	if err := tmplPostinstall.Execute(postinstall, payload); err != nil {
		return fmt.Errorf("could not generate postinstall script: %w", err)
	}

	pkg, err := macospkg.GeneratePkg("com.github.korylprince.macos-device-attestation", "1.0.0", postinstall.Bytes())
	if err != nil {
		return fmt.Errorf("could not generate payload pkg: %w", err)
	}

	signedPkg, err := macospkg.SignPkg(pkg, m.cert, m.key)
	if err != nil {
		return fmt.Errorf("could not sign payload pkg: %w", err)
	}

	fsPath, err := m.Put("payload.pkg", signedPkg)
	if err != nil {
		return fmt.Errorf("could not store payload pkg: %w", err)
	}

	manifest := macospkg.NewManifest(signedPkg, fmt.Sprintf("%s/%s", m.CachePrefix, fsPath), macospkg.ManifestHashSHA256)

	if err = m.InstallEnterpriseApplication(udid, manifest); err != nil {
		return fmt.Errorf("could not install payload: %w", err)
	}

	if err = m.InstallProfile(udid, profile); err != nil {
		return fmt.Errorf("could not install profile: %w", err)
	}

	return nil
}
