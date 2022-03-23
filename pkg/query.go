package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmppQueryReqPktLen  = HeaderPktLen + 2
	SmppQueryRespPktLen = HeaderPktLen + 2
)

type SmppQueryReqPkt struct {
	MsgID         string
	SourceAddrTON uint8  // 源地址编码类型
	SourceAddrNPI uint8  // 源地址编码方案
	SourceAddr    string // 提交该短消息的SME的地址

	// used in session
	SequenceNum uint32
}

func (p *SmppQueryReqPkt) Pack(seqId uint32) ([]byte, error) {
	msgId := NewCOctetString(p.MsgID).Byte(65)
	sourceAddr := NewCOctetString(p.SourceAddr).Byte(21)
	var commandLength = SmppQueryReqPktLen + uint32(len(msgId)) + uint32(len(sourceAddr))

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_QUERY),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	// body
	w.WriteBytes(msgId)
	w.WriteByte(p.SourceAddrTON)
	w.WriteByte(p.SourceAddrNPI)
	w.WriteBytes(sourceAddr)

	return w.Bytes()
}

func (p *SmppQueryReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	p.MsgID = string(r.ReadOCString(65))
	p.SourceAddrTON = r.ReadByte()
	p.SourceAddrNPI = r.ReadByte()
	p.SourceAddr = string(r.ReadOCString(21))

	return r.Error()
}

func (p *SmppQueryReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Query Req ---")
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "SourceAddrTON: ", p.SourceAddrTON)
	fmt.Fprintln(&b, "SourceAddrNPI: ", p.SourceAddrNPI)
	fmt.Fprintln(&b, "SourceAddr: ", p.SourceAddr)
	return b.String()
}

type SmppQueryRespPkt struct {
	MsgID        string
	FinalDate    string
	MessageState uint8
	ErrorCode    uint8

	// used in session
	Status      Status
	SequenceNum uint32
}

func (p *SmppQueryRespPkt) Pack(seqId uint32) ([]byte, error) {
	msgId := NewCOctetString(p.MsgID).Byte(65)
	finalDate := NewCOctetString(p.FinalDate).FixedByte(17)
	var commandLength = SmppQueryRespPktLen + uint32(len(msgId)) + uint32(len(finalDate))

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_QUERY_RESP),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	// body
	w.WriteBytes(msgId)
	w.WriteBytes(finalDate)
	w.WriteByte(p.MessageState)
	w.WriteByte(p.ErrorCode)

	return w.Bytes()
}

func (p *SmppQueryRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	p.MsgID = string(r.ReadOCString(65))
	p.FinalDate = string(r.ReadOCString(17))
	p.MessageState = r.ReadByte()
	p.ErrorCode = r.ReadByte()

	return r.Error()
}

func (p *SmppQueryRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Query Resp ---")
	fmt.Fprintln(&b, "Status: ", p.Status)
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "FinalDate: ", p.FinalDate)
	fmt.Fprintln(&b, "MessageState: ", p.MessageState)
	fmt.Fprintln(&b, "ErrorCode: ", p.ErrorCode)
	return b.String()
}
