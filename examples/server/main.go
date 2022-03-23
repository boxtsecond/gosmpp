package main

import (
	"log"
	"net"
	"time"

	"github.com/boxtsecond/gosmpp/pkg"
	"github.com/boxtsecond/gosmpp/server"
)

const (
	user     string = "100"
	password string = "12345678"
	spId     string = "123456"
)

func handleLogin(r *server.Response, p *server.Packet, l *log.Logger) (bool, error) {
	req, ok := p.Packer.(*pkg.SmppBindTransceiverReqPkt)
	if !ok {
		return true, nil
	}

	l.Println("remote addr:", p.Conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String())
	resp := r.Packer.(*pkg.SmppBindTransceiverRespPkt)

	if req.SystemID != user {
		resp.Status = pkg.ESME_RINVSYSID
		l.Println("handleLogin SystemID error:", resp.Status.Error())
		return false, resp.Status.Error()
	}

	if req.Password != password {
		resp.Status = pkg.ESME_RINVPASWD
		l.Println("handleLogin Password error:", resp.Status.Error())
		return false, resp.Status.Error()
	}

	resp.Status = pkg.ESME_ROK
	resp.SystemID = req.SystemID
	resp.ScInterfaceVersion = pkg.NewTLV(pkg.TAG_SCInterfaceVersion, []byte{pkg.VERSION})
	return false, nil
}

func handleSubmit(r *server.Response, p *server.Packet, l *log.Logger) (bool, error) {
	req, ok := p.Packer.(*pkg.SmppSubmitReqPkt)
	if !ok {
		return true, nil
	}

	resp := r.Packer.(*pkg.SmppSubmitRespPkt)
	resp.MsgID, _ = pkg.GenMsgID(spId, <-p.Conn.SequenceNum)
	deliverPkgs := make([]*pkg.SmppDeliverReqPkt, 0)
	l.Printf("handleSubmit: handle submit from %s ok! msgid[%s], destTerminalId[%s]\n",
		req.SourceAddr, resp.MsgID, req.DestinationAddr)

	t := pkg.GenNowTimeYYStr()
	msgStat := pkg.SmppDeliverMsgContent{
		SubmitMsgID: resp.MsgID,
		Sub:         "001",
		Dlvrd:       "001",
		SubmitDate:  t,
		DoneDate:    t,
		Stat:        "DELIVRD",
		Err:         "000",
		Txt:         "00000000000000000000",
	}
	msgContent := msgStat.Encode()
	deliverPkgs = append(deliverPkgs, &pkg.SmppDeliverReqPkt{
		ServiceType:        req.ServiceType,
		DestinationAddr:    req.DestinationAddr,
		PriorityFlag:       pkg.NORMAL_PRIORITY,
		RegisteredDelivery: 0,
		DataCoding:         pkg.ASCII,
		EsmClass:           pkg.SM_DELIVER,
		SmLength:           uint8(len(msgContent)),
		ShortMessage:       msgContent,
		SequenceNum:        <-p.Conn.SequenceNum,
	})
	go mockDeliver(deliverPkgs, p)
	return false, nil
}

func mockDeliver(pkgs []*pkg.SmppDeliverReqPkt, s *server.Packet) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:

			for _, p := range pkgs {
				err := s.SendPkt(p, p.SequenceNum)
				if err != nil {
					log.Printf("server smpp: send a smpp deliver request error: %s.", err)
					return
				} else {
					log.Printf("server smpp: send a smpp deliver request ok.")
				}
			}
			return

		default:
		}

	}
}

func main() {
	var handlers = []server.Handler{
		server.HandlerFunc(handleLogin),
		server.HandlerFunc(handleSubmit),
	}

	err := server.ListenAndServe(":8890",
		pkg.VERSION,
		5*time.Second,
		3,
		nil,
		handlers...,
	)
	if err != nil {
		log.Println("smpp Listen And Server error:", err)
	}
	return
}
