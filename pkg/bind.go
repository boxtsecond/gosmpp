package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmppBindTransceiverReqPktLen = HeaderPktLen + 1 + 1 + 1
)

type SmppBindTransceiverReqPkt struct {
	SystemID         string
	Password         string
	SystemType       string
	InterfaceVersion uint8
	AddrTON          uint8
	AddrNPI          uint8
	AddressRange     string
	// used in session
	SequenceNum uint32
}

func (p *SmppBindTransceiverReqPkt) Pack(seqId uint32) ([]byte, error) {
	systemId := NewCOctetString(p.SystemID).Byte(16)
	password := NewCOctetString(p.Password).Byte(9)
	systemType := NewCOctetString(p.SystemType).Byte(13)
	addressRange := NewCOctetString(p.AddressRange).Byte(41)

	commandLength := uint32(int(SmppBindTransceiverReqPktLen) + len(systemId) + len(password) + len(systemType) + len(addressRange))

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_BIND_TRANSCEIVER),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	// body
	w.WriteBytes(systemId)
	w.WriteBytes(password)
	w.WriteBytes(systemType)
	w.WriteByte(p.InterfaceVersion)
	w.WriteByte(p.AddrTON)
	w.WriteByte(p.AddrNPI)
	w.WriteBytes(addressRange)
	return w.Bytes()
}

func (p *SmppBindTransceiverReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	p.SystemID = string(r.ReadOCString(16))
	p.Password = string(r.ReadOCString(9))
	p.SystemType = string(r.ReadOCString(13))
	p.InterfaceVersion = r.ReadByte()
	p.AddrTON = r.ReadByte()
	p.AddrNPI = r.ReadByte()
	p.AddressRange = string(r.ReadOCString(16))
	return r.Error()
}

func (p *SmppBindTransceiverReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Bind Transceiver Req ---")
	fmt.Fprintln(&b, "SystemID: ", p.SystemID)
	fmt.Fprintln(&b, "Password: ", p.Password)
	fmt.Fprintln(&b, "SystemType: ", p.SystemType)
	fmt.Fprintln(&b, "InterfaceVersion: ", p.InterfaceVersion)
	fmt.Fprintln(&b, "AddrTON: ", p.AddrTON)
	fmt.Fprintln(&b, "AddrNPI: ", p.AddrNPI)
	fmt.Fprintln(&b, "AddressRange: ", p.AddressRange)
	return b.String()
}

type SmppBindTransceiverRespPkt struct {
	SystemID           string
	ScInterfaceVersion *TLV

	// used in session
	Status      Status // 请求返回结果
	SequenceNum uint32
}

func (p *SmppBindTransceiverRespPkt) Pack(seqId uint32) ([]byte, error) {
	systemId := NewCOctetString(p.SystemID).Byte(16)
	scInterfaceVersion, err := p.ScInterfaceVersion.Byte()
	if err != nil {
		return nil, err
	}
	commandLength := HeaderPktLen + uint32(len(systemId)) + uint32(p.ScInterfaceVersion.Len())

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_BIND_TRANSCEIVER_RESP),
		CommandStatus: uint32(p.Status),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)

	// body
	w.WriteBytes(systemId)
	w.WriteBytes(scInterfaceVersion)

	return w.Bytes()
}

func (p *SmppBindTransceiverRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)
	systemId := r.ReadOCString(16)
	p.SystemID = string(systemId)

	if len(systemId)+1 > len(data) {
		return r.Error()
	}
	options, err := ParseOptions(data[len(systemId)+1:])
	if err != nil {
		return err
	}
	p.ScInterfaceVersion = options[TAG_SCInterfaceVersion]
	return r.Error()
}

func (p *SmppBindTransceiverRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Bind Transceiver Resp ---")
	fmt.Fprintln(&b, "Status: ", p.Status)
	fmt.Fprintln(&b, "SystemID: ", p.SystemID)
	fmt.Fprintln(&b, "ScInterfaceVersion: ", p.ScInterfaceVersion)
	return b.String()
}
