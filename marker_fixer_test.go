package gortsplib

import (
	"testing"
	"time"

	"github.com/pion/rtp"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph264"
	"github.com/bluenviron/gortsplib/v5/pkg/format/rtph265"
)

// H265 RTP payloads. The byte after the NALU header carries first_slice_segment_in_pic_flag.
var (
	h265ParamsAP     = []byte{0x60, 0x01, 0x00, 0x03, 0x40, 0x01, 0x0c, 0x00, 0x03, 0x42, 0x01, 0x01}
	h265IDRFirst     = []byte{0x26, 0x01, 0xaf, 0x01, 0x02}
	h265IDRNext      = []byte{0x26, 0x01, 0x20, 0x01, 0x02}
	h265TrailFirst   = []byte{0x02, 0x01, 0xd0, 0x01, 0x02}
	h265TrailNext    = []byte{0x02, 0x01, 0x40, 0x01, 0x02}
	h265TrailFUFirst = []byte{0x62, 0x01, 0x81, 0xa0, 0x01, 0x02}
	h265TrailFUNext  = []byte{0x62, 0x01, 0x81, 0x20, 0x01, 0x02}
	h265TrailFUMid   = []byte{0x62, 0x01, 0x01, 0x03, 0x04}
	h265TrailFUEnd   = []byte{0x62, 0x01, 0x41, 0x05, 0x06}
	h265PrefixSEI    = []byte{0x4e, 0x01, 0x05, 0x01, 0x02}
)

// H264 RTP payloads. The byte after the NALU header carries first_mb_in_slice.
var (
	h264ParamsSTAPA = []byte{0x78, 0x00, 0x02, 0x67, 0x42, 0x00, 0x02, 0x68, 0xce}
	h264IDRFirst    = []byte{0x65, 0x88, 0x01, 0x02}
	h264IDRNext     = []byte{0x65, 0x40, 0x01, 0x02}
	h264NonIDRFirst = []byte{0x41, 0x9a, 0x01, 0x02}
	h264FUAStart    = []byte{0x5c, 0x81, 0x40, 0x01}
	h264FUAEnd      = []byte{0x5c, 0x41, 0x02, 0x03}
	h264SEI         = []byte{0x06, 0x05, 0x01, 0x02}
)

type testPacket struct {
	payload []byte
	marker  bool
}

// testPackets builds packets with increasing sequence numbers, one timestamp per picture.
func testPackets(pictures ...[]testPacket) []*rtp.Packet {
	var pkts []*rtp.Packet
	seq := uint16(100)

	for i, picture := range pictures {
		for _, p := range picture {
			pkts = append(pkts, &rtp.Packet{
				Header: rtp.Header{
					Version:        2,
					Marker:         p.marker,
					PayloadType:    96,
					SequenceNumber: seq,
					Timestamp:      uint32(i) * 3000,
					SSRC:           0x9dbb7812,
				},
				Payload: p.payload,
			})
			seq++
		}
	}

	return pkts
}

// h265PerSliceMarkerPictures is an IDR picture and three non-IDR pictures with three slices each,
// with the marker bit set at the end of every slice.
func h265PerSliceMarkerPictures() []*rtp.Packet {
	nonIDR := []testPacket{
		{h265TrailFUFirst, false},
		{h265TrailFUMid, false},
		{h265TrailFUEnd, true},
		{h265TrailNext, true},
		{h265TrailNext, true},
	}

	return testPackets(
		[]testPacket{
			{h265ParamsAP, false},
			{h265IDRFirst, true},
			{h265IDRNext, true},
			{h265IDRNext, true},
		},
		nonIDR,
		nonIDR,
		nonIDR,
		[]testPacket{{h265TrailFirst, true}},
	)
}

func markersByTimestamp(pkts []*rtp.Packet) map[uint32][]bool {
	ret := make(map[uint32][]bool)
	for _, pkt := range pkts {
		ret[pkt.Timestamp] = append(ret[pkt.Timestamp], pkt.Marker)
	}
	return ret
}

func TestStartsWithContinuationSlice(t *testing.T) {
	for _, ca := range []struct {
		name    string
		fn      func([]byte) bool
		payload []byte
		ret     bool
	}{
		{"h265 single first", h265StartsWithContinuationSlice, h265TrailFirst, false},
		{"h265 single next", h265StartsWithContinuationSlice, h265TrailNext, true},
		{"h265 idr next", h265StartsWithContinuationSlice, h265IDRNext, true},
		{"h265 fu start first", h265StartsWithContinuationSlice, h265TrailFUFirst, false},
		{"h265 fu start next", h265StartsWithContinuationSlice, h265TrailFUNext, true},
		{"h265 fu middle", h265StartsWithContinuationSlice, h265TrailFUMid, false},
		{"h265 fu end", h265StartsWithContinuationSlice, h265TrailFUEnd, false},
		{"h265 aggregation of parameters", h265StartsWithContinuationSlice, h265ParamsAP, false},
		{
			"h265 aggregation starting with slice",
			h265StartsWithContinuationSlice,
			[]byte{0x60, 0x01, 0x00, 0x03, 0x02, 0x01, 0x40},
			true,
		},
		{"h265 sei", h265StartsWithContinuationSlice, h265PrefixSEI, false},
		{"h265 too short", h265StartsWithContinuationSlice, []byte{0x02, 0x01}, false},
		{"h264 single first", h264StartsWithContinuationSlice, h264IDRFirst, false},
		{"h264 single next", h264StartsWithContinuationSlice, h264IDRNext, true},
		{"h264 non-idr first", h264StartsWithContinuationSlice, h264NonIDRFirst, false},
		{"h264 fu-a start next", h264StartsWithContinuationSlice, h264FUAStart, true},
		{"h264 fu-a end", h264StartsWithContinuationSlice, h264FUAEnd, false},
		{"h264 stap-a of parameters", h264StartsWithContinuationSlice, h264ParamsSTAPA, false},
		{
			"h264 stap-a starting with slice",
			h264StartsWithContinuationSlice,
			[]byte{0x78, 0x00, 0x02, 0x65, 0x40},
			true,
		},
		{"h264 sei", h264StartsWithContinuationSlice, h264SEI, false},
		{"h264 too short", h264StartsWithContinuationSlice, []byte{0x65}, false},
	} {
		t.Run(ca.name, func(t *testing.T) {
			require.Equal(t, ca.ret, ca.fn(ca.payload))
		})
	}
}

func TestMarkerFixerUnchanged(t *testing.T) {
	for _, ca := range []struct {
		name  string
		forma format.Format
		pkts  []*rtp.Packet
	}{
		{
			"h265 multiple slices",
			&format.H265{},
			testPackets(
				[]testPacket{{h265ParamsAP, false}, {h265IDRFirst, false}, {h265IDRNext, false}, {h265IDRNext, true}},
				[]testPacket{{h265TrailFUFirst, false}, {h265TrailFUMid, false}, {h265TrailFUEnd, false}, {h265TrailNext, true}},
				[]testPacket{{h265TrailFirst, false}, {h265TrailNext, true}},
			),
		},
		{
			"h265 trailing sei",
			&format.H265{},
			testPackets(
				[]testPacket{{h265IDRFirst, true}, {h265PrefixSEI, true}},
				[]testPacket{{h265TrailFirst, true}, {h265PrefixSEI, true}},
			),
		},
		{
			"h264 multiple slices",
			&format.H264{},
			testPackets(
				[]testPacket{{h264ParamsSTAPA, false}, {h264IDRFirst, false}, {h264IDRNext, true}},
				[]testPacket{{h264NonIDRFirst, false}, {h264FUAStart, false}, {h264FUAEnd, true}},
			),
		},
		{
			"h264 trailing sei",
			&format.H264{},
			testPackets(
				[]testPacket{{h264IDRFirst, true}, {h264SEI, true}},
				[]testPacket{{h264NonIDRFirst, true}, {h264SEI, true}},
			),
		},
	} {
		t.Run(ca.name, func(t *testing.T) {
			f := newMarkerFixer(ca.forma)
			require.NotNil(t, f)

			var out []*rtp.Packet

			for i, pkt := range ca.pkts {
				marker := pkt.Marker
				f.process(pkt, func(pkt *rtp.Packet) {
					out = append(out, pkt)
				})
				require.Len(t, out, i+1)
				require.Same(t, pkt, out[i])
				require.Equal(t, marker, out[i].Marker)
			}

			require.False(t, f.enabled)
		})
	}
}

func TestMarkerFixerNotAvailable(t *testing.T) {
	require.Nil(t, newMarkerFixer(&format.Opus{}))
	require.Nil(t, newMarkerFixer(&format.MJPEG{}))
}

func TestMarkerFixerH265(t *testing.T) {
	f := newMarkerFixer(&format.H265{})

	var out []*rtp.Packet

	for _, pkt := range h265PerSliceMarkerPictures() {
		f.process(pkt, func(pkt *rtp.Packet) {
			out = append(out, pkt)
		})
	}

	require.True(t, f.enabled)

	// the last packet waits for the next one
	require.NotNil(t, f.pending)
	require.Equal(t, uint32(4*3000), f.pending.Timestamp)

	for i := 1; i < len(out); i++ {
		require.Equal(t, out[i-1].SequenceNumber+1, out[i].SequenceNumber)
	}

	// the marker bit of the first slice of the first picture is sent before the issue is detected
	require.Equal(t, map[uint32][]bool{
		0:        {false, true, false, true},
		3000:     {false, false, false, false, true},
		2 * 3000: {false, false, false, false, true},
		3 * 3000: {false, false, false, false, true},
	}, markersByTimestamp(out))

	dec := &rtph265.Decoder{}
	err := dec.Init()
	require.NoError(t, err)

	var aus [][][]byte

	for _, pkt := range out {
		au, err2 := dec.Decode(pkt)
		if err2 == nil {
			aus = append(aus, au)
		}
	}

	require.Equal(t, [][][]byte{
		{
			{0x40, 0x01, 0x0c},
			{0x42, 0x01, 0x01},
			h265IDRFirst,
		},
		{h265IDRNext, h265IDRNext},
		{{0x02, 0x01, 0xa0, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, h265TrailNext, h265TrailNext},
		{{0x02, 0x01, 0xa0, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, h265TrailNext, h265TrailNext},
		{{0x02, 0x01, 0xa0, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, h265TrailNext, h265TrailNext},
	}, aus)
}

func TestMarkerFixerH264(t *testing.T) {
	f := newMarkerFixer(&format.H264{})

	var out []*rtp.Packet

	for _, pkt := range testPackets(
		[]testPacket{{h264ParamsSTAPA, false}, {h264IDRFirst, true}, {h264IDRNext, true}},
		[]testPacket{{h264NonIDRFirst, true}, {h264FUAStart, false}, {h264FUAEnd, true}},
		[]testPacket{{h264NonIDRFirst, true}, {h264FUAStart, false}, {h264FUAEnd, true}},
		[]testPacket{{h264NonIDRFirst, true}},
	) {
		f.process(pkt, func(pkt *rtp.Packet) {
			out = append(out, pkt)
		})
	}

	require.True(t, f.enabled)

	require.Equal(t, map[uint32][]bool{
		0:        {false, true, true},
		3000:     {false, false, true},
		2 * 3000: {false, false, true},
	}, markersByTimestamp(out))

	dec := &rtph264.Decoder{}
	err := dec.Init()
	require.NoError(t, err)

	var aus [][][]byte

	for _, pkt := range out {
		au, err2 := dec.Decode(pkt)
		if err2 == nil {
			aus = append(aus, au)
		}
	}

	require.Equal(t, [][][]byte{
		{{0x67, 0x42}, {0x68, 0xce}, h264IDRFirst},
		{h264IDRNext},
		{h264NonIDRFirst, {0x41, 0x40, 0x01, 0x02, 0x03}},
		{h264NonIDRFirst, {0x41, 0x40, 0x01, 0x02, 0x03}},
	}, aus)
}

func TestClientPlayPerSliceMarker(t *testing.T) {
	medi := &description.Media{
		Type:    description.MediaTypeVideo,
		Formats: []format.Format{&format.H265{PayloadTyp: 96}},
	}

	var stream *ServerStream

	s := &Server{
		Handler: &testServerHandler{
			onDescribe: func(_ *ServerHandlerOnDescribeCtx) (*base.Response, *ServerStream, error) {
				return &base.Response{StatusCode: base.StatusOK}, stream, nil
			},
			onSetup: func(_ *ServerHandlerOnSetupCtx) (*base.Response, *ServerStream, error) {
				return &base.Response{StatusCode: base.StatusOK}, stream, nil
			},
			onPlay: func(_ *ServerHandlerOnPlayCtx) (*base.Response, error) {
				return &base.Response{StatusCode: base.StatusOK}, nil
			},
		},
		RTSPAddress: "localhost:8554",
	}

	err := s.Start()
	require.NoError(t, err)
	defer s.Close()

	stream = &ServerStream{
		Server: s,
		Desc:   &description.Session{Medias: []*description.Media{medi}},
	}
	err = stream.Initialize()
	require.NoError(t, err)
	defer stream.Close()

	tcp := ProtocolTCP
	c := Client{Scheme: "rtsp", Host: "localhost:8554", Protocol: &tcp}

	err = c.Start()
	require.NoError(t, err)
	defer c.Close()

	desc, _, err := c.Describe(mustParseURL("rtsp://localhost:8554/teststream"))
	require.NoError(t, err)

	err = c.SetupAll(desc.BaseURL, desc.Medias)
	require.NoError(t, err)

	pkts := h265PerSliceMarkerPictures()
	recv := make(chan *rtp.Packet, len(pkts))

	c.OnPacketRTPAny(func(_ *description.Media, _ format.Format, pkt *rtp.Packet) {
		recv <- pkt
	})

	_, err = c.Play(nil)
	require.NoError(t, err)

	for _, pkt := range pkts {
		err = stream.WritePacketRTP(medi, pkt)
		require.NoError(t, err)
	}

	var out []*rtp.Packet

	for range len(pkts) - 1 {
		select {
		case pkt := <-recv:
			out = append(out, pkt)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out")
		}
	}

	for i, pkt := range out {
		require.Equal(t, pkts[i].Payload, pkt.Payload)
	}

	require.Equal(t, map[uint32][]bool{
		0:        {false, true, false, true},
		3000:     {false, false, false, false, true},
		2 * 3000: {false, false, false, false, true},
		3 * 3000: {false, false, false, false, true},
	}, markersByTimestamp(out))
}

func TestServerRecordPerSliceMarker(t *testing.T) {
	pkts := h265PerSliceMarkerPictures()
	recv := make(chan *rtp.Packet, len(pkts))

	s := &Server{
		Handler: &testServerHandler{
			onAnnounce: func(_ *ServerHandlerOnAnnounceCtx) (*base.Response, error) {
				return &base.Response{StatusCode: base.StatusOK}, nil
			},
			onSetup: func(_ *ServerHandlerOnSetupCtx) (*base.Response, *ServerStream, error) {
				return &base.Response{StatusCode: base.StatusOK}, nil, nil
			},
			onRecord: func(ctx *ServerHandlerOnRecordCtx) (*base.Response, error) {
				ctx.Session.OnPacketRTPAny(func(_ *description.Media, _ format.Format, pkt *rtp.Packet) {
					recv <- pkt
				})
				return &base.Response{StatusCode: base.StatusOK}, nil
			},
		},
		RTSPAddress: "localhost:8554",
	}

	err := s.Start()
	require.NoError(t, err)
	defer s.Close()

	desc := &description.Session{Medias: []*description.Media{{
		Type:    description.MediaTypeVideo,
		Formats: []format.Format{&format.H265{PayloadTyp: 96}},
	}}}

	tcp := ProtocolTCP
	c := Client{Protocol: &tcp}

	err = c.StartRecording("rtsp://localhost:8554/teststream", desc)
	require.NoError(t, err)
	defer c.Close()

	for _, pkt := range pkts {
		err = c.WritePacketRTP(desc.Medias[0], pkt)
		require.NoError(t, err)
	}

	var out []*rtp.Packet

	for range len(pkts) - 1 {
		select {
		case pkt := <-recv:
			out = append(out, pkt)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out")
		}
	}

	for i, pkt := range out {
		require.Equal(t, pkts[i].Payload, pkt.Payload)
	}

	require.Equal(t, map[uint32][]bool{
		0:        {false, true, false, true},
		3000:     {false, false, false, false, true},
		2 * 3000: {false, false, false, false, true},
		3 * 3000: {false, false, false, false, true},
	}, markersByTimestamp(out))
}
