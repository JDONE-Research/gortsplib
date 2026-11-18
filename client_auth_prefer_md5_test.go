package gortsplib

import (
	"bufio"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bluenviron/gortsplib/v5/pkg/auth"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/conn"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/headers"
)

// a server that offers both Digest SHA-256 and Digest MD5 but accepts only MD5
// (like some camera firmwares) must be authenticated at the first attempt.
func TestClientAuthDigestPreferMD5(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:8554")
	require.NoError(t, err)
	defer l.Close()

	serverDone := make(chan struct{})
	defer func() { <-serverDone }()

	go func() {
		defer close(serverDone)

		nconn, err2 := l.Accept()
		require.NoError(t, err2)
		defer nconn.Close()
		conn := conn.NewConn(bufio.NewReader(nconn), nconn)

		req, err2 := conn.ReadRequest()
		require.NoError(t, err2)
		require.Equal(t, base.Options, req.Method)

		err2 = conn.WriteResponse(&base.Response{
			StatusCode: base.StatusOK,
			Header: base.Header{
				"Public": base.HeaderValue{string(base.Describe)},
			},
		})
		require.NoError(t, err2)

		req, err2 = conn.ReadRequest()
		require.NoError(t, err2)
		require.Equal(t, base.Describe, req.Method)

		nonce, err2 := auth.GenerateNonce()
		require.NoError(t, err2)

		err2 = conn.WriteResponse(&base.Response{
			StatusCode: base.StatusUnauthorized,
			Header: base.Header{
				"WWW-Authenticate": auth.GenerateWWWAuthenticate(
					[]auth.VerifyMethod{auth.VerifyMethodDigestSHA256, auth.VerifyMethodDigestMD5},
					"IPCAM", nonce),
			},
		})
		require.NoError(t, err2)

		req, err2 = conn.ReadRequest()
		require.NoError(t, err2)
		require.Equal(t, base.Describe, req.Method)

		var authz headers.Authorization
		err2 = authz.Unmarshal(req.Header["Authorization"])
		require.NoError(t, err2)
		require.NotNil(t, authz.Algorithm)
		require.Equal(t, headers.AuthAlgorithmMD5, *authz.Algorithm)

		err2 = auth.Verify(req, "myuser", "mypass",
			[]auth.VerifyMethod{auth.VerifyMethodDigestMD5}, "IPCAM", nonce)
		require.NoError(t, err2)

		err2 = conn.WriteResponse(&base.Response{
			StatusCode: base.StatusOK,
			Header: base.Header{
				"Content-Type": base.HeaderValue{"application/sdp"},
			},
			Body: mediasToSDP([]*description.Media{testH264Media}),
		})
		require.NoError(t, err2)
	}()

	u, err := base.ParseURL("rtsp://myuser:mypass@localhost:8554/stream")
	require.NoError(t, err)

	c := Client{
		Scheme: u.Scheme,
		Host:   u.Host,
	}

	err = c.Start()
	require.NoError(t, err)
	defer c.Close()

	_, _, err = c.Describe(u)
	require.NoError(t, err)
}
