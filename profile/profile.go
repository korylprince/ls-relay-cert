package profile

import (
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Config is used to specify defaults for the generated profile
type Config struct {
	// PayloadVersion is the version of the inner payload
	PayloadVersion int
	// PayloadIdentifier is used as the top level identifier, and a prefix for the inner payload
	PayloadIdentifier string
	// PayloadUUID is used as the top level UUID
	PayloadUUID         string
	PayloadOrganization string
}

type TopLevelProfile struct {
	// PayloadType is the payload type, specified on each payload domain's reference page.
	PayloadType string

	// PayloadVersion is the version of this specific payload.
	PayloadVersion int

	// PayloadIdentifier is the reverse-DNS-style identifier for the payload. This identifier is usually the same as the TopLevel value, with an additional component appended.
	PayloadIdentifier string

	// PayloadUUID is the globally unique identifier for the payload. The actual content is unimportant, but must be globally unique. In macOS, use uuidgen to generate UUIDs.
	PayloadUUID string

	// PayloadDisplayName is the human-readable name for the profile payload. The name is displayed on the Detail screen and doesn't have to be unique.
	PayloadDisplayName string

	// PayloadDescription is the human-readable description of this payload. This description is shown on the Detail screen.
	PayloadDescription string

	// PayloadOrganization is the human-readable string containing the name of the organization that provided the profile. This value doesn't need to match the organization payload value in the enclosing dictionary.
	PayloadOrganization string

	// PayloadScope is a string that defines whether the profile should be installed for the system or the user. In many cases, it determines the location of certificate items, such as keychains.
	PayloadScope string

	// PayloadRemovalDisallowed, if present and set to true, the user cannot delete the profile (unless the profile has a removal password and the user provides it).
	PayloadRemovalDisallowed bool

	// PayloadContent contains one or more CertificateRootProfiles
	PayloadContent []*CertificateRootProfile
}

type CertificateRootProfile struct {
	// PayloadType is the payload type, specified on each payload domain's reference page.
	PayloadType string

	// PayloadVersion is the version of this specific payload.
	PayloadVersion int

	// PayloadIdentifier is the reverse-DNS-style identifier for the payload. This identifier is usually the same as the TopLevel value, with an additional component appended.
	PayloadIdentifier string

	// PayloadUUID is the globally unique identifier for the payload. The actual content is unimportant, but must be globally unique. In macOS, use uuidgen to generate UUIDs.
	PayloadUUID string

	// PayloadDisplayName is the human-readable name for the profile payload. The name is displayed on the Detail screen and doesn't have to be unique.
	PayloadDisplayName string

	// PayloadOrganization is the human-readable string containing the name of the organization that provided the profile. This value doesn't need to match the organization payload value in the enclosing dictionary.
	PayloadOrganization string

	// PayloadContent is the binary representation of the payload encoded in base64
	PayloadContent []byte
}

// New returns a profile for the given ca certificate
func New(config *Config, ca *x509.Certificate) (*TopLevelProfile, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("could not generate uuid: %w", err)
	}

	p := &TopLevelProfile{
		PayloadType:              "Configuration",
		PayloadVersion:           1,
		PayloadIdentifier:        config.PayloadIdentifier,
		PayloadUUID:              config.PayloadUUID,
		PayloadDisplayName:       "Lightspeed Relay Smart Agent",
		PayloadDescription:       "Root Certificate for Lightspeed Relay Smart Agent",
		PayloadOrganization:      config.PayloadOrganization,
		PayloadScope:             "System",
		PayloadRemovalDisallowed: false,
		PayloadContent: []*CertificateRootProfile{{
			PayloadType:         "com.apple.security.root",
			PayloadVersion:      config.PayloadVersion,
			PayloadIdentifier:   config.PayloadIdentifier + ".root-certificate",
			PayloadUUID:         strings.ToUpper(u.String()),
			PayloadDisplayName:  "Root Certificate",
			PayloadOrganization: config.PayloadOrganization,
			PayloadContent:      ca.Raw,
		}},
	}

	return p, nil
}
