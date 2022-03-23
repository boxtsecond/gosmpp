package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmppUnbindReqPktLen  = HeaderPktLen
	SmppUnbindRespPktLen = HeaderPktLen
)

type SmppUnbindReqPkt struct {
	// used in session
	SequenceNum uint32
}
type SmppUnbindRespPkt struct {
	// used in session
	Status      Status
	SequenceNum uint32
}

func (p *SmppUnbindReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmppUnbindReqPktLen)

	// header
	header := Header{
		CommandLength: SmppUnbindReqPktLen,
		CommandID:     uint32(SMPP_UNBIND),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	return w.Bytes()
}

func (p *SmppUnbindReqPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmppUnbindReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Unbind Req ---")
	return b.String()
}

func (p *SmppUnbindRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmppUnbindRespPktLen)

	// header
	header := Header{
		CommandLength: SmppUnbindRespPktLen,
		CommandID:     uint32(SMPP_UNBIND_RESP),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	return w.Bytes()
}

func (p *SmppUnbindRespPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmppUnbindRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Unbind Resp ---")
	return b.String()
}
