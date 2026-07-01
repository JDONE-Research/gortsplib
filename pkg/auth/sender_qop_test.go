package auth

import (
	"regexp"
	"testing"

	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/stretchr/testify/require"
)

// verifies the RFC 7616 qop="auth" path end-to-end:
// Unmarshal captures qop -> AddAuthorization takes the qop branch ->
// Marshal emits qop/nc/cnonce and a response consistent with
// H(HA1:nonce:nc:cnonce:qop:HA2).
func TestSenderDigestQOPAuth(t *testing.T) {
	for _, ca := range []struct {
		name  string
		extra string
		hash  func(string) string
	}{
		{"sha256", `, algorithm="SHA-256", qop="auth"`, sha256Hex},
		{"md5", `, qop="auth"`, md5Hex},
		{"qop list", `, algorithm="SHA-256", qop="auth,auth-int"`, sha256Hex},
	} {
		t.Run(ca.name, func(t *testing.T) {
			se := &Sender{
				WWWAuth: base.HeaderValue{
					`Digest realm="myrealm", nonce="f49ac6dd0ba708d4becddc9692d1f2ce"` + ca.extra,
				},
				User: "myuser",
				Pass: "mypass",
			}
			require.NoError(t, se.Initialize())

			req := &base.Request{
				Method: base.Setup,
				URL:    mustParseURL("rtsp://myhost/mypath?key=val/trackID=3"),
			}
			se.AddAuthorization(req)

			got := req.Header["Authorization"][0]
			require.Contains(t, got, "qop=auth")
			require.Contains(t, got, "nc=00000001")
			require.Regexp(t, `cnonce="[0-9a-f]{32}"`, got)

			cnonce := regexp.MustCompile(`cnonce="([0-9a-f]+)"`).FindStringSubmatch(got)[1]
			ha1 := ca.hash("myuser:myrealm:mypass")
			ha2 := ca.hash("SETUP:rtsp://myhost/mypath?key=val/trackID=3")
			want := ca.hash(ha1 + ":f49ac6dd0ba708d4becddc9692d1f2ce:00000001:" + cnonce + ":auth:" + ha2)
			require.Contains(t, got, `response="`+want+`"`)
		})
	}
}

// without qop, the sender must keep the legacy RFC 2069 response and emit no qop.
func TestSenderDigestNoQOPUnchanged(t *testing.T) {
	se := &Sender{
		WWWAuth: base.HeaderValue{
			`Digest realm="myrealm", nonce="f49ac6dd0ba708d4becddc9692d1f2ce"`,
		},
		User: "myuser",
		Pass: "mypass",
	}
	require.NoError(t, se.Initialize())

	req := &base.Request{
		Method: base.Setup,
		URL:    mustParseURL("rtsp://myhost/mypath?key=val/trackID=3"),
	}
	se.AddAuthorization(req)

	got := req.Header["Authorization"][0]
	require.NotContains(t, got, "qop")
	require.NotContains(t, got, "cnonce")
	require.Contains(t, got, `response="ba6e9cccbfeb38db775378a0a9067ba5"`)
}
