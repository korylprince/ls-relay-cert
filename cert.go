package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// GenerateCA generates a new CA key pair
func GenerateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return nil, nil, fmt.Errorf("could not generate serial: %w", err)
	}

	serial := new(big.Int)
	serial.SetBytes(buf)

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate key: %w", err)
	}

	ski := sha512.Sum512(x509.MarshalPKCS1PublicKey(&key.PublicKey))

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:    "Lightspeed Filter Agent",
			Organization:  []string{"Lightspeed Systems"},
			Country:       []string{"US"},
			Locality:      []string{"Austin"},
			Province:      []string{"Texas"},
			StreetAddress: []string{"2500 Bee Cave Road, Suite 350"},
			PostalCode:    []string{"78746"},
		},
		SubjectKeyId:          ski[:],
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse generated certificate: %w", err)
	}

	return cert, key, nil
}

// GenerateLocalhost generates a new key pair for localhost signed by the given CA key pair
func GenerateLocalhost(ca *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return nil, nil, fmt.Errorf("could not generate serial: %w", err)
	}

	serial := new(big.Int)
	serial.SetBytes(buf)

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate key: %w", err)
	}

	ski := sha512.Sum512(x509.MarshalPKCS1PublicKey(&key.PublicKey))

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Lightspeed Systems"},
		},
		SubjectKeyId:          ski[:],
		DNSNames:              []string{"localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse generated certificate: %w", err)
	}

	return cert, key, nil
}

// GeneratePKI generates and returns a certificate root profile and PEM encoded CA and localhost key pairs
func GeneratePKI(config *ProfileConfig) (*TopLevelProfile, *Payload, error) {
	c, ck, err := GenerateCA()
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate CA key pair: %w", err)
	}

	lh, lhk, err := GenerateLocalhost(c, ck)
	if err != nil {
		return nil, nil, fmt.Errorf("could not generate localhost key pair: %w", err)
	}

	profile, err := NewProfile(config, c)
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
