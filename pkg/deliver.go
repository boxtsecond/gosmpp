package pkg

import (
	"bytes"
	"fmt"
)

const (
	SmppDeliverRespPktLen = HeaderPktLen + 1
)

type SmppDeliverMsgContent struct {
	SubmitMsgID string // submit resp 的 MsgID
	Sub         string
	Dlvrd       string
	SubmitDate  string
	DoneDate    string
	Stat        string
	Err         string
	Txt         string
}

func (p *SmppDeliverMsgContent) Encode() []byte {
	p.Dlvrd = NewOctetString(p.Dlvrd).FixedString(3)
	p.SubmitDate = NewOctetString(p.SubmitDate).FixedString(10)
	p.DoneDate = NewOctetString(p.DoneDate).FixedString(10)
	p.Stat = NewOctetString(p.Stat).FixedString(7)
	p.Err = NewOctetString(p.Err).FixedString(3)
	var b bytes.Buffer
	b.WriteString("id:")
	b.WriteString(p.SubmitMsgID)
	b.WriteString(" sub:")
	b.WriteString(p.Sub)
	b.WriteString(" dlvrd:")
	b.WriteString(p.Dlvrd)
	b.WriteString(" submit date:")
	b.WriteString(p.SubmitDate)
	b.WriteString(" done date:")
	b.WriteString(p.DoneDate)
	b.WriteString(" stat:")
	b.WriteString(p.Stat)
	b.WriteString(" err:")
	b.WriteString(p.Err)
	b.WriteString(" text:")
	b.WriteString(p.Txt)

	return b.Bytes()
}

func DecodeDeliverMsgContent(data []byte) *SmppDeliverMsgContent {
	p := &SmppDeliverMsgContent{}
	// 标准协议长度
	var rr = newPkgReader(data)
	rr.ReadBytes([]byte("id:"))
	p.SubmitMsgID = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("sub:"))
	p.Sub = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("dlvrd:"))
	p.Dlvrd = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("submit date:"))
	p.SubmitDate = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("done date:"))
	p.DoneDate = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("stat:"))
	p.Stat = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("err:"))
	p.Err = string(rr.ReadOCStringBySpace())
	rr.ReadBytes([]byte("text:"))
	if rr.Len() != 0 {
		p.Txt = string(rr.ReadCString(20))
	}

	if rr.Error() != nil {
		var r = newPkgReader(data)
		r.ReadBytes([]byte("id:"))
		p.SubmitMsgID = string(r.ReadOCString(65))
		r.ReadBytes([]byte(" sub:"))
		p.Sub = string(r.ReadOCString(3))
		r.ReadBytes([]byte(" dlvrd:"))
		p.Dlvrd = string(r.ReadCString(3))
		r.ReadBytes([]byte(" submit date:"))
		p.SubmitDate = string(r.ReadCString(10))
		r.ReadBytes([]byte(" done date:"))
		p.DoneDate = string(r.ReadCString(10))
		r.ReadBytes([]byte(" stat:"))
		p.Stat = string(r.ReadCString(7))
		r.ReadBytes([]byte(" err:"))
		p.Err = string(r.ReadCString(3))
		r.ReadBytes([]byte(" text:"))
		p.Txt = string(r.ReadCString(20))
	}

	return p
}

func (p *SmppDeliverMsgContent) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "\tID(SubmitMsgID): ", p.SubmitMsgID)
	fmt.Fprintln(&b, "\tSub: ", p.Sub)
	fmt.Fprintln(&b, "\tDlvrd: ", p.Dlvrd)
	fmt.Fprintln(&b, "\tSubmitDate: ", p.SubmitDate)
	fmt.Fprintln(&b, "\tDoneDate: ", p.DoneDate)
	fmt.Fprintln(&b, "\tStat: ", p.Stat)
	fmt.Fprintln(&b, "\tErr: ", p.Err)
	fmt.Fprintln(&b, "\tTxt: ", p.Txt)

	return b.String()
}

type SmppDeliverReqPkt struct {
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
	ScheduleDeliveryTime string // 必须为 NULL
	ValidityPeriod       string // 必须为 NULL
	RegisteredDelivery   uint8  // 标识 SMSC 是否要状态 报告或 SME 是否要确认标识
	ReplaceIfPresentFlag uint8  // 替换现存短消息标志
	DataCoding           uint8  // 短消息用户数据编码方案
	SmDefaultMsgID       uint8  // 预定义短消息 ID
	SmLength             uint8  // 短消息长度
	ShortMessage         string // 短消息内容

	// 可选字段
	Options Options

	// used in session
	SequenceNum    uint32
	MsgStatContent *SmppDeliverMsgContent
}

func (p *SmppDeliverReqPkt) Pack(seqId uint32) ([]byte, error) {
	serviceType := NewCOctetString(p.ServiceType).Byte(6)
	sourceAddr := NewCOctetString(p.SourceAddr).Byte(21)
	destinationAddr := NewCOctetString(p.DestinationAddr).Byte(21)
	scheduleDeliveryTime := NewCOctetString("").FixedByte(1)
	validityPeriod := scheduleDeliveryTime
	content := NewOctetString(p.ShortMessage).Bytes(254)
	p.SmLength = uint8(len(content))

	var commandLength = uint32(int(HeaderPktLen) + 12 + len(serviceType) + len(sourceAddr) + len(destinationAddr) + len(scheduleDeliveryTime) + len(validityPeriod) + len(content) + p.Options.Len())

	var w = newPkgWriter(commandLength)
	// header
	header := Header{
		CommandLength: commandLength,
		CommandID:     uint32(SMPP_DELIVER),
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

func (p *SmppDeliverReqPkt) Unpack(data []byte) error {
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
	scheduleDeliveryTime := r.ReadOCString(1)
	p.ScheduleDeliveryTime = string(scheduleDeliveryTime)
	validityPeriod := r.ReadOCString(1)
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

func (p *SmppDeliverReqPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Deliver Req ---")
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

type SmppDeliverRespPkt struct {
	MsgID string

	// used in session
	Status      Status
	SequenceNum uint32
}

func (p *SmppDeliverRespPkt) Pack(seqId uint32) ([]byte, error) {
	var w = newPkgWriter(SmppDeliverRespPktLen)
	// header
	header := Header{
		CommandLength: SmppDeliverRespPktLen,
		CommandID:     uint32(SMPP_DELIVER_RESP),
		SequenceNum:   seqId,
	}
	w.WriteHeader(header)
	p.SequenceNum = seqId

	// body
	w.WriteBytes(NewCOctetString("").FixedByte(1))
	return w.Bytes()
}

func (p *SmppDeliverRespPkt) Unpack(data []byte) error {
	var r = newPkgReader(data)

	// Body: MsgID
	msgId := r.ReadOCString(9)
	p.MsgID = string(msgId)
	return r.Error()
}

func (p *SmppDeliverRespPkt) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- SMPP Deliver Resp ---")
	fmt.Fprintln(&b, "MsgID: ", p.MsgID)
	fmt.Fprintln(&b, "Status: ", p.Status)
	return b.String()
}
