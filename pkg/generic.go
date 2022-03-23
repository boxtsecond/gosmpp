package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmppGenericNackReqPktLen = HeaderPktLen
)

type SmppGenericNackReqPkt struct {
	// used in session
	Status      Status
	SequenceNum uint32
}

func (p *SmppGenericNackReqPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmppGenericNackReqPktLen)

	// header
	header := Header{
		CommandLength: SmppGenericNackReqPktLen,
		CommandID:     uint32(SMPP_GENERIC_NACK),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	return w.Bytes()
}

func (p *SmppGenericNackReqPkt) Unpack(data []byte) error {
	return nil
}

func (p *SmppGenericNackReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP GenericNack Req ---")
	return b.String()
}
