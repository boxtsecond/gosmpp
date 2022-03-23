package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const HeaderPktLen uint32 = 4 + 4 + 4 + 4

// 消息头(所有消息公共包头)
type Header struct {
	CommandLength uint32 // Command_length 字段定义了整个 SMPP 数据包的长度
	CommandID     uint32 // SMPP PDU 消息类型
	CommandStatus uint32 // Command_status 命令状态字段表示请求消息是否成功
	SequenceNum   uint32 // 消息的序列号
}

func (p *Header) Pack(w *pkgWriter, pktLen, commandID, commandStatus, seqNum uint32) *pkgWriter {
	w.WriteInt(binary.BigEndian, pktLen)
	w.WriteInt(binary.BigEndian, commandID)
	w.WriteInt(binary.BigEndian, commandStatus)
	w.WriteInt(binary.BigEndian, seqNum)
	return w
}

func (p *Header) Unpack(r *pkgReader) *Header {
	r.ReadInt(binary.BigEndian, &p.CommandLength)
	r.ReadInt(binary.BigEndian, &p.CommandID)
	r.ReadInt(binary.BigEndian, &p.CommandStatus)
	r.ReadInt(binary.BigEndian, &p.SequenceNum)
	return p
}

func (p *Header) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "--- Header ---")
	fmt.Fprintln(&b, "CommandLength: ", p.CommandLength)
	fmt.Fprintf(&b, "CommandID: 0x%x\n", p.CommandID)
	fmt.Fprintln(&b, "CommandStatus: ", p.CommandStatus)
	fmt.Fprintln(&b, "SequenceNum: ", p.SequenceNum)

	return b.String()

}
