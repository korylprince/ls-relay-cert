package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/groob/plist"
	"github.com/korylprince/ls-relay-cert/mdm"
	"github.com/korylprince/ls-relay-cert/profile"
)

func main() {
	flVersion := flag.Int("version", 1, "The version used for the profile")
	flIdentifier := flag.String("identifier", "com.github.korylprince.ls-relay-cert", "The top level profile identifier, and a prefix for the inner payload")
	flUUID := flag.String("uuid", "randomly generated", "The UUID used for the profile")
	flOrg := flag.String("org", "Lightspeed Systems", "The organization used for the profile")
	flYears := flag.Int("years", 10, "The number of years to use for the CA")
	flOutput := flag.String("out", ".", "Output directory")
	flag.Parse()

	if *flUUID == "randomly generated" {
		*flUUID = strings.ToUpper(uuid.NewString())
	}

	prof, certs, err := mdm.GeneratePKI(*flYears, &profile.Config{
		PayloadVersion:      *flVersion,
		PayloadIdentifier:   *flIdentifier,
		PayloadUUID:         *flUUID,
		PayloadOrganization: *flOrg,
	})

	if err != nil {
		fmt.Println("could not generate profile and certificates:", err)
		os.Exit(-1)
	}

	// marshal profile
	buf, err := plist.MarshalIndent(prof, "\t")
	if err != nil {
		fmt.Println("could not marshal plist:", err)
		os.Exit(-1)
	}

	// write profile
	if err := os.WriteFile(filepath.Join(*flOutput, "Lightspeed Certificate.mobileconfig"), buf, 0644); err != nil {
		fmt.Println("could not write profile:", err)
		os.Exit(-1)
	}

	// write ca cert
	if err := os.WriteFile(filepath.Join(*flOutput, "ca.pem"), []byte(certs.CA), 0644); err != nil {
		fmt.Println("could not write ca.pem:", err)
		os.Exit(-1)
	}

	// write ca key
	if err := os.WriteFile(filepath.Join(*flOutput, "ca_key.pem"), []byte(certs.CAKey), 0600); err != nil {
		fmt.Println("could not write ca_key.pem:", err)
		os.Exit(-1)
	}

	// write localhost cert
	if err := os.WriteFile(filepath.Join(*flOutput, "localhost.pem"), []byte(certs.Localhost), 0644); err != nil {
		fmt.Println("could not write localhost.pem:", err)
		os.Exit(-1)
	}

	// write localhost key
	if err := os.WriteFile(filepath.Join(*flOutput, "localhost_key.pem"), []byte(certs.LocalhostKey), 0600); err != nil {
		fmt.Println("could not write localhost_key.pem:", err)
		os.Exit(-1)
	}
}
