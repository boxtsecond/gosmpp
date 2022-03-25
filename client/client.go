package client

import (
	"errors"
	"net"
	"time"

	"github.com/boxtsecond/gosmpp/pkg"
)

var ErrRespNotMatch = errors.New("the response is not matched with the request")

type Client struct {
	conn *pkg.Conn
	ver  uint8
}

func NewClient(version uint8) *Client {
	return &Client{
		ver: version,
	}
}

func (cli *Client) Connect(serverAddr, systemID, password, systemType string, addrTON, addrNPI uint8, addrRange string, timeout time.Duration) error {
	var err error
	conn, err := net.DialTimeout("tcp", serverAddr, timeout)
	if err != nil {
		return err
	}
	cli.conn = pkg.NewConnection(conn, cli.ver)
	defer func() {
		if err != nil {
			if cli.conn != nil {
				cli.conn.Close()
			}
		}
	}()

	cli.conn.SetState(pkg.CONNECTION_CONNECTED)

	// Login to the server.
	req := &pkg.SmppBindTransceiverReqPkt{
		SystemID:         systemID,
		Password:         password,
		SystemType:       systemType,
		InterfaceVersion: cli.ver,
		AddrTON:          addrTON,
		AddrNPI:          addrNPI,
		AddressRange:     addrRange,
	}

	_, err = cli.SendReqPkt(req)
	if err != nil {
		return err
	}

	p, err := cli.conn.RecvAndUnpackPkt(timeout)
	if err != nil {
		return err
	}

	rsp, ok := p.(*pkg.SmppBindTransceiverRespPkt)
	if !ok {
		err = ErrRespNotMatch
		return err
	}

	if rsp.Status.Data() != 0 {
		return rsp.Status.Error()
	}

	cli.conn.SetState(pkg.CONNECTION_AUTHOK)
	return nil
}

func (cli *Client) Disconnect() {
	if cli.conn != nil {
		cli.conn.Close()
	}
}

func (cli *Client) SendReqPkt(packet pkg.Packer) (uint32, error) {
	seq := <-cli.conn.SequenceNum
	return seq, cli.conn.SendPkt(packet, seq)
}

func (cli *Client) SendRspPkt(packet pkg.Packer, sequenceID uint32) error {
	return cli.conn.SendPkt(packet, sequenceID)
}

func (cli *Client) RecvAndUnpackPkt(timeout time.Duration) (interface{}, error) {
	return cli.conn.RecvAndUnpackPkt(timeout)
}

func (cli *Client) GetConn() *pkg.Conn {
	return cli.conn
}
