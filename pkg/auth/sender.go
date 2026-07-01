package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/headers"
)

// Sender allows to send credentials.
type Sender struct {
	WWWAuth base.HeaderValue
	User    string
	Pass    string

	authHeader *headers.Authenticate
}

// Initialize initializes a Sender.
func (se *Sender) Initialize() error {
	for _, v := range se.WWWAuth {
		var auth headers.Authenticate
		err := auth.Unmarshal(base.HeaderValue{v})
		if err != nil {
			continue // ignore unrecognized headers
		}

		if se.authHeader == nil ||
			(auth.Algorithm != nil && *auth.Algorithm == headers.AuthAlgorithmSHA256) ||
			(se.authHeader.Method == headers.AuthMethodBasic) {
			se.authHeader = &auth
		}
	}

	if se.authHeader == nil {
		return fmt.Errorf("no authentication methods available")
	}

	return nil
}

// AddAuthorization adds the Authorization header to a Request.
func (se *Sender) AddAuthorization(req *base.Request) {
	urStr := req.URL.CloneWithoutCredentials().String()

	h := headers.Authorization{
		Method: se.authHeader.Method,
	}

	h.Username = se.User

	if se.authHeader.Method == headers.AuthMethodBasic {
		h.BasicPass = se.Pass
	} else { // digest
		h.Realm = se.authHeader.Realm
		h.Nonce = se.authHeader.Nonce
		h.URI = urStr
		h.Algorithm = se.authHeader.Algorithm
		h.Opaque = se.authHeader.Opaque

		hash := md5Hex
		if se.authHeader.Algorithm != nil && *se.authHeader.Algorithm == headers.AuthAlgorithmSHA256 {
			hash = sha256Hex
		}

		ha1 := hash(se.User + ":" + se.authHeader.Realm + ":" + se.Pass)
		ha2 := hash(string(req.Method) + ":" + urStr)

		if se.authHeader.QOP != nil && hasAuthQOP(*se.authHeader.QOP) {
			cnonce := generateCnonce()
			nc := "00000001"
			h.QOP = "auth"
			h.Cnonce = cnonce
			h.Nc = nc
			// RFC 7616: response = H(HA1:nonce:nc:cnonce:qop:HA2)
			h.Response = hash(ha1 + ":" + se.authHeader.Nonce + ":" +
				nc + ":" + cnonce + ":auth:" + ha2)
		} else {
			// RFC 2069 (no qop): response = H(HA1:nonce:HA2)
			h.Response = hash(ha1 + ":" + se.authHeader.Nonce + ":" + ha2)
		}
	}

	if req.Header == nil {
		req.Header = make(base.Header)
	}

	req.Header["Authorization"] = h.Marshal()
}

// generateCnonce returns a random client nonce.
func generateCnonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// hasAuthQOP reports whether the qop challenge offers "auth".
func hasAuthQOP(qop string) bool {
	for _, v := range strings.Split(qop, ",") {
		if strings.TrimSpace(v) == "auth" {
			return true
		}
	}
	return false
}
