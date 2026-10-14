package gortsplib

import (
	"github.com/pion/rtp"

	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

// h264IsContinuationSlice checks whether a NALU is a slice that is not the first slice of a picture.
func h264IsContinuationSlice(typ h264.NALUType, sliceHeader []byte) bool {
	if typ != h264.NALUTypeNonIDR && typ != h264.NALUTypeIDR {
		return false
	}

	// first_mb_in_slice is coded with ue(v), therefore its first bit is set only when it is zero.
	return len(sliceHeader) != 0 && (sliceHeader[0]&0x80) == 0
}

// h264StartsWithContinuationSlice checks whether a RTP/H264 payload starts with a continuation slice.
func h264StartsWithContinuationSlice(payload []byte) bool {
	if len(payload) < 2 {
		return false
	}

	switch typ := h264.NALUType(payload[0] & 0x1F); typ {
	case h264.NALUTypeSTAPA:
		if len(payload) < 5 {
			return false
		}
		return h264IsContinuationSlice(h264.NALUType(payload[3]&0x1F), payload[4:])

	case h264.NALUTypeFUA:
		if len(payload) < 3 || (payload[1]&0x80) == 0 {
			return false
		}
		return h264IsContinuationSlice(h264.NALUType(payload[1]&0x1F), payload[2:])

	default:
		return h264IsContinuationSlice(typ, payload[1:])
	}
}

// h265IsContinuationSlice checks whether a NALU is a slice segment that is not the first one of a picture.
func h265IsContinuationSlice(typ h265.NALUType, sliceHeader []byte) bool {
	if typ >= h265.NALUType_VPS_NUT {
		return false
	}

	// first_slice_segment_in_pic_flag
	return len(sliceHeader) != 0 && (sliceHeader[0]&0x80) == 0
}

// h265StartsWithContinuationSlice checks whether a RTP/H265 payload starts with a continuation slice.
func h265StartsWithContinuationSlice(payload []byte) bool {
	if len(payload) < 3 {
		return false
	}

	switch typ := h265.NALUType((payload[0] >> 1) & 0b111111); typ {
	case h265.NALUType_AggregationUnit:
		if len(payload) < 7 {
			return false
		}
		return h265IsContinuationSlice(h265.NALUType((payload[4]>>1)&0b111111), payload[6:])

	case h265.NALUType_FragmentationUnit:
		if len(payload) < 4 || (payload[2]&0x80) == 0 {
			return false
		}
		return h265IsContinuationSlice(h265.NALUType(payload[2]&0b111111), payload[3:])

	default:
		return h265IsContinuationSlice(typ, payload[2:])
	}
}

// markerFixer fixes H264 and H265 streams that set the RTP marker bit at the end of every slice,
// instead of at the end of every access unit (RFC6184, RFC7798).
// Depacketizers end access units at the marker bit, therefore these streams produce access units
// that contain a single slice, and that are rejected by DTS extractors and recorders.
//
// The fixer is enabled when a continuation slice with the same timestamp is received after a marker bit.
// Then, packets with the marker bit are delayed until the next packet is received,
// and their marker bit is cleared when the next packet continues the same picture.
// Timestamps are not touched, therefore PTS and NTP of packets are preserved.
type markerFixer struct {
	startsWithContinuationSlice func([]byte) bool

	enabled       bool
	prevMarker    bool
	prevTimestamp uint32
	pending       *rtp.Packet
}

func newMarkerFixer(forma format.Format) *markerFixer {
	switch forma.(type) {
	case *format.H264:
		return &markerFixer{startsWithContinuationSlice: h264StartsWithContinuationSlice}

	case *format.H265:
		return &markerFixer{startsWithContinuationSlice: h265StartsWithContinuationSlice}
	}

	return nil
}

func (f *markerFixer) process(pkt *rtp.Packet, onPacketRTP OnPacketRTPFunc) {
	if !f.enabled {
		if f.prevMarker && pkt.Timestamp == f.prevTimestamp && f.startsWithContinuationSlice(pkt.Payload) {
			f.enabled = true
		} else {
			f.prevMarker = pkt.Marker
			f.prevTimestamp = pkt.Timestamp
			onPacketRTP(pkt)
			return
		}
	}

	if f.pending != nil {
		if pkt.Timestamp == f.pending.Timestamp && f.startsWithContinuationSlice(pkt.Payload) {
			f.pending.Marker = false
		}

		onPacketRTP(f.pending)
		f.pending = nil
	}

	if pkt.Marker {
		f.pending = pkt
		return
	}

	onPacketRTP(pkt)
}
