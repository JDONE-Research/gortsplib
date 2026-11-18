package auth

import (
	"testing"

	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/headers"
	"github.com/stretchr/testify/require"
)

// verifies that, among multiple challenges, Digest MD5 is preferred over
// Digest SHA-256 regardless of order, and Digest SHA-256 over Basic.
func TestSenderAlgorithmPreference(t *testing.T) {
	const (
		basic  = `Basic realm="myrealm"`
		md5    = `Digest realm="myrealm", nonce="f49ac6dd0ba708d4becddc9692d1f2ce", algorithm="MD5"`
		md5Imp = `Digest realm="myrealm", nonce="f49ac6dd0ba708d4becddc9692d1f2ce"`
		sha256 = `Digest realm="myrealm", nonce="f49ac6dd0ba708d4becddc9692d1f2ce", algorithm="SHA-256"`
	)

	md5Alg := headers.AuthAlgorithmMD5
	sha256Alg := headers.AuthAlgorithmSHA256

	for _, ca := range []struct {
		name      string
		wwwAuth   base.HeaderValue
		method    headers.AuthMethod
		algorithm *headers.AuthAlgorithm
	}{
		{"sha256 then md5", base.HeaderValue{sha256, md5}, headers.AuthMethodDigest, &md5Alg},
		{"md5 then sha256", base.HeaderValue{md5, sha256}, headers.AuthMethodDigest, &md5Alg},
		{"sha256 then implicit md5", base.HeaderValue{sha256, md5Imp}, headers.AuthMethodDigest, nil},
		{"basic sha256 md5", base.HeaderValue{basic, sha256, md5}, headers.AuthMethodDigest, &md5Alg},
		{"sha256 md5 basic", base.HeaderValue{sha256, md5, basic}, headers.AuthMethodDigest, &md5Alg},
		{"basic then sha256", base.HeaderValue{basic, sha256}, headers.AuthMethodDigest, &sha256Alg},
		{"sha256 then basic", base.HeaderValue{sha256, basic}, headers.AuthMethodDigest, &sha256Alg},
		{"sha256 only", base.HeaderValue{sha256}, headers.AuthMethodDigest, &sha256Alg},
		{"basic only", base.HeaderValue{basic}, headers.AuthMethodBasic, nil},
	} {
		t.Run(ca.name, func(t *testing.T) {
			se := &Sender{
				WWWAuth: ca.wwwAuth,
				User:    "myuser",
				Pass:    "mypass",
			}
			require.NoError(t, se.Initialize())
			require.Equal(t, ca.method, se.authHeader.Method)
			require.Equal(t, ca.algorithm, se.authHeader.Algorithm)
		})
	}
}
