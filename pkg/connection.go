package pkg

import (
	"encoding/binary"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

type State uint8

const (
	CONNECTION_CLOSED State = iota
	CONNECTION_CONNECTED
	CONNECTION_AUTHOK
)

type Conn struct {
	net.Conn
	State   State
	Version uint8

	// for SequenceNum generator goroutine
	SequenceNum <-chan uint32
	done        chan<- struct{}
}

// The sequence number may range from: 0x00000001 to 0x7FFFFFFF.
func newSequenceNumGenerator() (<-chan uint32, chan<- struct{}) {
	out := make(chan uint32)
	done := make(chan struct{})

	go func() {
		rand.Seed(time.Now().UnixNano())
		var i = uint32(rand.Intn(0x7FFFFFFF))
		if i == 0 {
			i = 1
		}
		for {
			select {
			case out <- i:
				i++
				if i == 0 || i > 0x7FFFFFFF {
					i = 1
				}
			case <-done:
				close(out)
				return
			}
		}
	}()
	return out, done
}

func NewConnection(conn net.Conn, v uint8) *Conn {
	sequenceNum, done := newSequenceNumGenerator()
	c := &Conn{
		Conn:        conn,
		Version:     v,
		SequenceNum: sequenceNum,
		done:        done,
	}
	tc := c.Conn.(*net.TCPConn)
	tc.SetKeepAlive(true) //Keepalive as default
	return c
}

func (c *Conn) Close() {
	if c != nil {
		if c.State == CONNECTION_CLOSED {
			return
		}
		close(c.done)  // let the SeqId goroutine exit.
		c.Conn.Close() // close the underlying net.Conn
		c.State = CONNECTION_CLOSED
	}
}

func (c *Conn) SetState(state State) {
	c.State = state
}

func (c *Conn) SendPkt(packet Packer, seqId uint32) error {
	if c.State == CONNECTION_CLOSED {
		return ErrConnIsClosed
	}

	data, err := packet.Pack(seqId)
	if err != nil {
		return err
	}

	_, err = c.Conn.Write(data) //block write
	if err != nil {
		return err
	}

	return nil
}

const (
	defaultReadBufferSize = 4096
)

type readBuffer struct {
	Header
	leftData [defaultReadBufferSize]byte
}

var readBufferPool = sync.Pool{
	New: func() interface{} {
		return &readBuffer{}
	},
}

func (c *Conn) RecvAndUnpackPkt(timeout time.Duration) (Packer, error) {
	if c.State == CONNECTION_CLOSED {
		return nil, ErrConnIsClosed
	}
	rb := readBufferPool.Get().(*readBuffer)
	defer func() {
		readBufferPool.Put(rb)
		c.SetReadDeadline(time.Time{})
	}()

	if timeout != 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}

	// packet header
	err := binary.Read(c.Conn, binary.BigEndian, &rb.Header)
	if err != nil {
		return nil, err
	}

	if rb.Header.CommandLength < SMPP_PACKET_MIN || rb.Header.CommandLength > SMPP_PACKET_MAX {
		return nil, ErrTotalLengthInvalid
	}

	if !((CommandID(rb.Header.CommandID) > SMPP_REQUEST_MIN && CommandID(rb.Header.CommandID) < SMPP_REQUEST_MAX) ||
		(CommandID(rb.Header.CommandID) >= SMPP_RESPONSE_MIN && CommandID(rb.Header.CommandID) < SMPP_RESPONSE_MAX)) {
		return nil, ErrCommandIDInvalid
	}

	if timeout != 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}

	// packet body
	var leftData = rb.leftData[0:(rb.Header.CommandLength - HeaderPktLen)]
	if len(leftData) > 0 {
		_, err = io.ReadFull(c.Conn, leftData)
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok {
				if netErr.Timeout() {
					return nil, ErrReadPktBodyTimeout
				}
			}
			return nil, err
		}
	}

	var p Packer
	sequenceNum := rb.Header.SequenceNum
	status := Status(rb.Header.CommandStatus)

	switch CommandID(rb.Header.CommandID) {
	case SMPP_ENQUIRE_LINK:
		p = &SmppEnquireLinkReqPkt{SequenceNum: sequenceNum}
	case SMPP_ENQUIRE_LINK_RESP:
		p = &SmppEnquireLinkRespPkt{SequenceNum: sequenceNum}
	case SMPP_BIND_TRANSCEIVER:
		p = &SmppBindTransceiverReqPkt{SequenceNum: sequenceNum}
	case SMPP_BIND_TRANSCEIVER_RESP:
		p = &SmppBindTransceiverRespPkt{SequenceNum: sequenceNum, Status: status}
	case SMPP_SUBMIT:
		p = &SmppSubmitReqPkt{SequenceNum: sequenceNum}
	case SMPP_SUBMIT_RESP:
		p = &SmppSubmitRespPkt{SequenceNum: sequenceNum, Status: status}
	case SMPP_DELIVER:
		p = &SmppDeliverReqPkt{SequenceNum: sequenceNum}
	case SMPP_DELIVER_RESP:
		p = &SmppDeliverRespPkt{SequenceNum: sequenceNum, Status: status}
	case SMPP_UNBIND:
		p = &SmppUnbindReqPkt{SequenceNum: sequenceNum}
	case SMPP_UNBIND_RESP:
		p = &SmppUnbindRespPkt{SequenceNum: sequenceNum, Status: status}
	case SMPP_GENERIC_NACK:
		p = &SmppGenericNackReqPkt{SequenceNum: sequenceNum, Status: status}
	case SMPP_QUERY:
		p = &SmppQueryReqPkt{SequenceNum: sequenceNum}
	case SMPP_QUERY_RESP:
		p = &SmppQueryRespPkt{SequenceNum: sequenceNum}

	default:
		return nil, ErrCommandIDNotSupported
	}

	err = p.Unpack(leftData)
	if err != nil {
		return nil, err
	}
	return p, nil
}
