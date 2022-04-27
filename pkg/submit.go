package pkg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type SmppSubmitReqPkt struct {
	ServiceType          string // 指示联系到 SMS 应用服务消息的类型
	SourceAddrTON        uint8  // 源地址编码类型
	SourceAddrNPI        uint8  // 源地址编码方案
	SourceAddr           string // 提交该短消息的SME的地址
	DestAddrTON          uint8  // 目的地址编码类型
	DestAddrNPI          uint8  // 目的地址编码方案
	DestinationAddr      string // 短消息的目的地址
	EsmClass             uint8  // 指定信息模式和信息类型
	ProtocolID           uint8  // 协议指示和网络标识区
	PriorityFlag         uint8  // 指示短消息的优先级
	ScheduleDeliveryTime string // 表示计划下发该短消息的时间 如立即发送设置为 NULL，长度 1 或 17
	ValidityPeriod       string // 表示短消息的最后生存期限 如果需要 SMSC 默认有效期 设置为 NULL，长度 1 或 17
	RegisteredDelivery   uint8  // 标识 SMSC 是否要状态 报告或 SME 是否要确认标识
	ReplaceIfPresentFlag uint8  // 替换现存短消息标志
	DataCoding           uint8  // 短消息用户数据编码方案
	SmDefaultMsgID       uint8  // 预定义短消息 ID
	SmLength             uint8  // 短消息长度
	ShortMessage         string // 短消息内容

	// 可选参数
	Options Options

	// used in session
	SequenceNum uint32
}

func (p *SmppSubmitReqPkt) Pack(seqId uint32) ([]byte, error) {
	serviceType := NewCOctetString(p.ServiceType).Byte(6)
	sourceAddr := NewCOctetString(p.SourceAddr).Byte(21)
	destinationAddr := NewCOctetString(p.DestinationAddr).Byte(21)
	scheduleDeliveryTime := NewCOctetString(p.ScheduleDeliveryTime).FixedByte(17)
	validityPeriod := NewCOctetString(p.ValidityPeriod).FixedByte(17)
	content := NewOctetString(p.ShortMessage).Bytes(254)
	p.SmLength = uint8(len(content))

	var commandLength = uint32(int(HeaderPktLen) + 12 + len(serviceType) + len(sourceAddr) + len(destinationAddr) + len(scheduleDeliveryTime) + len(validityPeriod) + len(content) + p.Options.Len())

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_SUBMIT),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	// body
	w.WriteBytes(serviceType)
	w.WriteByte(p.SourceAddrTON)
	w.WriteByte(p.SourceAddrNPI)
	w.WriteBytes(sourceAddr)
	w.WriteByte(p.DestAddrTON)
	w.WriteByte(p.DestAddrNPI)
	w.WriteBytes(destinationAddr)
	w.WriteByte(p.EsmClass)
	w.WriteByte(p.ProtocolID)
	w.WriteByte(p.PriorityFlag)
	w.WriteBytes(scheduleDeliveryTime)
	w.WriteBytes(validityPeriod)
	w.WriteByte(p.RegisteredDelivery)
	w.WriteByte(p.ReplaceIfPresentFlag)
	w.WriteByte(p.DataCoding)
	w.WriteByte(p.SmDefaultMsgID)
	w.WriteByte(p.SmLength)
	w.WriteBytes(content)

	for _, o := range p.Options {
		b, _ := o.Byte()
		w.WriteBytes(b)
	}

	return w.Bytes()
}

func (p *SmppSubmitReqPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	serviceType := r.ReadOCString(6)
	p.ServiceType = string(serviceType)
	p.SourceAddrTON = r.ReadByte()
	p.SourceAddrNPI = r.ReadByte()
	sourceAddr := r.ReadOCString(21)
	p.SourceAddr = string(sourceAddr)
	p.DestAddrTON = r.ReadByte()
	p.DestAddrNPI = r.ReadByte()
	destinationAddr := r.ReadOCString(21)
	p.DestinationAddr = string(destinationAddr)
	p.EsmClass = r.ReadByte()
	p.ProtocolID = r.ReadByte()
	p.PriorityFlag = r.ReadByte()
	scheduleDeliveryTime := r.ReadOCString(17)
	p.ScheduleDeliveryTime = string(scheduleDeliveryTime)
	validityPeriod := r.ReadOCString(17)
	p.ValidityPeriod = string(validityPeriod)
	p.RegisteredDelivery = r.ReadByte()
	p.ReplaceIfPresentFlag = r.ReadByte()
	p.DataCoding = r.ReadByte()
	p.SmDefaultMsgID = r.ReadByte()
	p.SmLength = r.ReadByte()
	msgContent := make([]byte, p.SmLength)
	r.ReadBytes(msgContent)
	p.ShortMessage = string(msgContent)

	offset := int(p.SmLength) + len(serviceType) + len(sourceAddr) + len(destinationAddr) + len(scheduleDeliveryTime) + len(validityPeriod) + 12 + 5
	options, err := ParseOptions(data[offset:])
	if err != nil {
		return err
	}
	p.Options = options

	return r.Error()
}

func (p *SmppSubmitReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Submit Req ---")
	fmt.Fprintln(&b, "ServiceType: ", p.ServiceType)
	fmt.Fprintln(&b, "SourceAddrTON: ", p.SourceAddrTON)
	fmt.Fprintln(&b, "SourceAddrNPI: ", p.SourceAddrNPI)
	fmt.Fprintln(&b, "SourceAddr: ", p.SourceAddr)
	fmt.Fprintln(&b, "DestAddrTON: ", p.DestAddrTON)
	fmt.Fprintln(&b, "DestAddrNPI: ", p.DestAddrNPI)
	fmt.Fprintln(&b, "DestinationAddr: ", p.DestinationAddr)
	fmt.Fprintln(&b, "EsmClass: ", p.EsmClass)
	fmt.Fprintln(&b, "ProtocolID: ", p.ProtocolID)
	fmt.Fprintln(&b, "PriorityFlag: ", p.PriorityFlag)
	fmt.Fprintln(&b, "ScheduleDeliveryTime: ", p.ScheduleDeliveryTime)
	fmt.Fprintln(&b, "ValidityPeriod: ", p.ValidityPeriod)
	fmt.Fprintln(&b, "RegisteredDelivery: ", p.RegisteredDelivery)
	fmt.Fprintln(&b, "ReplaceIfPresentFlag: ", p.ReplaceIfPresentFlag)
	fmt.Fprintln(&b, "DataCoding: ", p.DataCoding)
	fmt.Fprintln(&b, "SmDefaultMsgID: ", p.SmDefaultMsgID)
	fmt.Fprintln(&b, "SmLength: ", p.SmLength)
	fmt.Fprintln(&b, "ShortMessage: ", p.ShortMessage)
	fmt.Fprintln(&b, "Options: ", p.Options.String())

	return b.String()
}

type SmppSubmitContentHeaderReqPkg struct {
	LastProtocolLen uint8
	UniqueIdLen     uint8
	LastLen         uint8
	UniqueId        uint16 //6位header时 1byte 7位header时 2byte
	PkTotal         uint8
	PkNumber        uint8
}

func GetSubmitMsgHeader(msgContent []byte) (*SmppSubmitContentHeaderReqPkg, error) {
	header := &SmppSubmitContentHeaderReqPkg{}
	r := newPkgReader(msgContent)
	r.ReadInt(binary.BigEndian, &header.LastProtocolLen)
	r.ReadInt(binary.BigEndian, &header.UniqueIdLen)
	r.ReadInt(binary.BigEndian, &header.LastLen)

	// 获取长短信header信息
	if header.LastProtocolLen == 5 { // 6位协议头
		if header.LastLen == 3 && header.UniqueIdLen == 0 {
			uniqueId := make([]byte, 1)
			r.ReadBytes(uniqueId)
			header.UniqueId = uint16(uniqueId[0])
			r.ReadInt(binary.BigEndian, &header.PkTotal)
			r.ReadInt(binary.BigEndian, &header.PkNumber)
		}
	} else if header.LastProtocolLen == 6 { // 7位协议头
		if header.LastLen == 4 && header.UniqueIdLen == 8 {
			r.ReadInt(binary.BigEndian, &header.UniqueId)
			r.ReadInt(binary.BigEndian, &header.PkTotal)
			r.ReadInt(binary.BigEndian, &header.PkNumber)
		}
	} else {
		return header, errors.New("msg header len illegal")
	}

	if header.PkNumber == 0 || header.PkNumber > header.PkTotal {
		return header, errors.New("msg header len illegal")
	}

	return header, r.Error()
}

type SmppSubmitRespPkt struct {
	MsgID string

	// used in session
	Status      Status
	SequenceNum uint32
}

func (p *SmppSubmitRespPkt) Pack(seqId uint32) ([]byte, error) {
	msgId := NewCOctetString(p.MsgID).Byte(65)
	var commandLength = HeaderPktLen + uint32(len(msgId))

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_SUBMIT_RESP),
		CommandStatus: uint32(p.Status),
		SequenceNum:   seqId,
	}
	p.SequenceNum = seqId
	w.WriteHeader(header)
	// body
	w.WriteBytes(msgId)
	return w.Bytes()
}

func (p *SmppSubmitRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	// Body: MsgID
	msgId := r.ReadOCString(65)
	p.MsgID = string(msgId)
	return r.Error()
}

func (p *SmppSubmitRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Submit Resp ---")
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "Status: ", p.Status)
	return b.String()
}
