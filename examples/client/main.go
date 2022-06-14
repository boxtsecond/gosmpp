package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/boxtsecond/gosmpp/client"
	"github.com/boxtsecond/gosmpp/pkg"
)

var (
	addr       = flag.String("addr", "0.0.0.0:8890", "smpp addr(运营商地址)")
	systemID   = flag.String("systemID", "100", "登陆账号")
	password   = flag.String("password", "12345678", "登陆密码")
	systemType = flag.String("systemType", "", "系统类型")
	phone      = flag.String("phone", "8618012345678", "接收手机号码, 86...")
	msg        = flag.String("msg", "【优刻得】验证码：1234", "短信内容")
	//msg = flag.String("msg", "【闪送】您的订单取件密码为 000000 (请妥善保管),订单尾号0000,预计08/16 10:24上门取件。闪送员张师傅,电话8618012345678(本单已开启号码保护,请务必使用本机号码呼叫)。查看闪送员实时位置请点击http://a.bcdefghigk.com/42bag0xV。打开微信-发现-搜一搜-搜索“闪送”查看。", "短信内容")
)

func startAClient(idx int) {
	c := client.NewClient(pkg.VERSION)
	defer wg.Done()
	defer c.Disconnect()

	fmt.Println("============")
	fmt.Println("addr: ", *addr)
	fmt.Println("systemID: ", *systemID)
	fmt.Println("password: ", *password)
	fmt.Println("systemType: ", *systemType)
	fmt.Println("phone: ", *phone)
	err := c.Connect(*addr, *systemID, *password, *systemType, 0, 0, "", 3*time.Second)

	if err != nil {
		log.Printf("client %d: connect error: %s.", idx, err)
		return
	}
	log.Printf("client %d: connect and auth ok", idx)

	t := time.NewTicker(time.Second)
	defer t.Stop()
	maxSubmit := 1
	count := 0

	go func() {
		for {
			// recv packets
			i, err := c.RecvAndUnpackPkt(0)
			if err != nil {
				log.Printf("client %d: client read and unpack pkt error: %s.", idx, err)
				return
			}

			switch p := i.(type) {
			case *pkg.SmppSubmitRespPkt:
				log.Printf("client %d: receive a smpp submit response: \n%v", idx, p)

			case *pkg.SmppDeliverReqPkt:
				log.Printf("client %d: receive a smpp deliver request: \n%v", idx, p)
				if p.EsmClass == pkg.SM_DELIVER {
					log.Printf("client %d: the smpp deliver request: %s is a status report.", idx, p)
				}
				rsp := &pkg.SmppDeliverRespPkt{
					Status: pkg.Status(0),
				}

				err := c.SendRspPkt(rsp, p.SequenceNum)
				if err != nil {
					log.Printf("client %d: send smpp deliver response error: %s.", idx, err)
					break
				} else {
					log.Printf("client %d: send smpp deliver response ok.", idx)
				}

			case *pkg.SmppEnquireLinkReqPkt:
				log.Printf("client %d: receive a smpp active request.", idx)
				rsp := &pkg.SmppEnquireLinkRespPkt{}
				err := c.SendRspPkt(rsp, p.SequenceNum)
				if err != nil {
					log.Printf("client %d: send smpp active response error: %s.", idx, err)
					break
				}
			case *pkg.SmppEnquireLinkRespPkt:
				log.Printf("client %d: receive a smpp active response.", idx)

			case *pkg.SmppUnbindReqPkt:
				log.Printf("client %d: receive a smpp unbind request.", idx)
				rsp := &pkg.SmppUnbindRespPkt{}
				err := c.SendRspPkt(rsp, p.SequenceNum)
				if err != nil {
					log.Printf("client %d: send smpp unbind response error: %s.", idx, err)
					break
				}
			case *pkg.SmppUnbindRespPkt:
				log.Printf("client %d: receive a smpp unbind response.", idx)
			}
		}
	}()

	for {
		select {
		case <-t.C:
			if count >= maxSubmit {
				continue
			}

			p := &pkg.SmppSubmitReqPkt{
				DestinationAddr:    *phone,
				EsmClass:           0,
				PriorityFlag:       pkg.NORMAL_PRIORITY,
				RegisteredDelivery: pkg.NEED_REPORT,
				DataCoding:         pkg.ASCII,
				ShortMessage:       *msg,
			}

			pkgs, err := pkg.GetMsgPkgs(p)
			if err != nil {
				log.Printf("client %d: get long msg pkg error: %s.", idx, err)
				continue
			}

			for _, req := range pkgs {
				_, err = c.SendReqPkt(req)
			}
			if err != nil {
				log.Printf("client %d: send a smpp submit request error: %s.", idx, err)
				return
			} else {
				log.Printf("client %d: send a smpp submit request ok", idx)
			}
			count += 1
		default:
		}
	}
}

var wg sync.WaitGroup

func init() {
	flag.Parse()
}

func main() {
	log.Println("Client example start!")
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go startAClient(i + 1)
	}

	wg.Wait()
	log.Println("Client example ends!")
}
