package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmppEnquireLinkReqPktLen  = HeaderPktLen
	SmppEnquireLinkRespPktLen = HeaderPktLen
)

type SmppEnquireLinkReqPkt struct {
	// used in session
	SequenceNum uint32
}
type SmppEnquireLinkRespPkt struct {
	// used in session
	SequenceNum uint32
}

func (p *SmppEnquireLinkReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmppEnquireLinkReqPktLen)

	// header
	header := Header{
		CommandLength: SmppEnquireLinkReqPktLen,
		CommandID:     uint32(SMPP_ENQUIRE_LINK),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	return w.Bytes()
}

func (p *SmppEnquireLinkReqPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmppEnquireLinkReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP EnquireLink Req ---")
	return b.String()
}

func (p *SmppEnquireLinkRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmppEnquireLinkRespPktLen)

	// header
	header := Header{
		CommandLength: SmppEnquireLinkRespPktLen,
		CommandID:     uint32(SMPP_ENQUIRE_LINK_RESP),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	return w.Bytes()
}

func (p *SmppEnquireLinkRespPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmppEnquireLinkRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP EnquireLink Resp ---")
	return b.String()
}
