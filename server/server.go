package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/boxtsecond/gosmpp/pkg"
)

var (
	ErrEmptyServerAddr    = errors.New("smpp server listen: empty server addr")
	ErrNoHandlers         = errors.New("smpp server: no connection handler")
	ErrUnsupportedPkt     = errors.New("smpp server read packet: receive a unsupported pkt")
	ErrUnsupportedVersion = errors.New("smpp server read packet: receive a unsupported version")
)

type Packet struct {
	pkg.Packer
	*pkg.Conn
}

type Response struct {
	*Packet
	pkg.Packer
	SequenceNum uint32
}

type Handler interface {
	ServeSmpp(*Response, *Packet) (bool, error)
}

type HandlerFunc func(*Response, *Packet) (bool, error)

func (f HandlerFunc) ServeSmpp(r *Response, p *Packet) (bool, error) {
	return f(r, p)
}

type Server struct {
	Addr    string
	Handler Handler

	// protocol info
	Version     uint8
	ReadTimeout time.Duration
	T           time.Duration
	N           int32

	ErrorLog *log.Logger
}

type conn struct {
	*pkg.Conn
	server      *Server
	readTimeout time.Duration

	// for enquire link
	t       time.Duration // interval between two enquire links
	n       int32         // continuous send times when no response back
	done    chan struct{}
	exceed  chan struct{}
	counter int32
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration
	for {
		rw, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.ErrorLog.Printf("accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		c, err := srv.newConn(rw)
		if err != nil {
			continue
		}

		srv.ErrorLog.Printf("accept a connection from %v\n", c.Conn.RemoteAddr())
		go c.serve()
	}
}

func (c *conn) readPacket() (*Response, error) {
	readTimeout := c.readTimeout
	i, err := c.Conn.RecvAndUnpackPkt(readTimeout)
	if err != nil {
		return nil, err
	}
	ver := c.server.Version

	var rsp *Response
	switch p := i.(type) {
	case *pkg.SmppBindTransceiverReqPkt:
		if p.InterfaceVersion != ver {
			return nil, pkg.NewOpError(ErrUnsupportedVersion,
				fmt.Sprintf("readPacket: receive unsupported version: %#v", p))
		}
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmppBindTransceiverRespPkt{
				SequenceNum: p.SequenceNum,
			},
			SequenceNum: p.SequenceNum,
		}
		c.server.ErrorLog.Printf("receive a smpp bind transceiver request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppSubmitReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmppSubmitRespPkt{
				SequenceNum: p.SequenceNum,
			},
			SequenceNum: p.SequenceNum,
		}
		c.server.ErrorLog.Printf("receive a smpp submit request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppDeliverReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmppDeliverRespPkt{
				SequenceNum: p.SequenceNum,
			},
			SequenceNum: p.SequenceNum,
		}
		c.server.ErrorLog.Printf("receive a smpp deliver response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppDeliverRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smpp deliver response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppEnquireLinkReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmppEnquireLinkRespPkt{
				SequenceNum: p.SequenceNum,
			},
			SequenceNum: p.SequenceNum,
		}
		c.server.ErrorLog.Printf("receive a smpp active request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppEnquireLinkRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smpp active response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppUnbindReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmppUnbindRespPkt{
				SequenceNum: p.SequenceNum,
			},
			SequenceNum: p.SequenceNum,
		}
		c.server.ErrorLog.Printf("receive a smpp exit request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppUnbindRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smpp exit response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppQueryReqPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
			Packer: &pkg.SmppQueryRespPkt{
				SequenceNum: p.SequenceNum,
			},
			SequenceNum: p.SequenceNum,
		}
		c.server.ErrorLog.Printf("receive a smpp query request from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	case *pkg.SmppQueryRespPkt:
		rsp = &Response{
			Packet: &Packet{
				Packer: p,
				Conn:   c.Conn,
			},
		}
		c.server.ErrorLog.Printf("receive a smpp query response from %v[%d]\n",
			c.Conn.RemoteAddr(), p.SequenceNum)

	default:
		return nil, pkg.NewOpError(ErrUnsupportedPkt,
			fmt.Sprintf("readPacket: receive unsupported packet type: %#v", p))
	}
	return rsp, nil
}

func (c *conn) close() {
	p := &pkg.SmppUnbindReqPkt{}

	err := c.Conn.SendPkt(p, <-c.Conn.SequenceNum)
	if err != nil {
		c.server.ErrorLog.Printf("send smpp exit request packet to %v error: %v\n", c.Conn.RemoteAddr(), err)
	}

	close(c.done)
	c.server.ErrorLog.Printf("close connection with %v!\n", c.Conn.RemoteAddr())
	c.Conn.Close()
}

func (c *conn) finishPacket(r *Response) error {
	//if _, ok := r.Packet.Packer.(*pkg.SmppEnquireLinkRespPkt); ok {
	//	atomic.AddInt32(&c.counter, -1)
	//	return nil
	//}

	if r.Packet != nil {
		atomic.StoreInt32(&c.counter, 0)
	}

	if r.Packer == nil {
		return nil
	}

	return c.Conn.SendPkt(r.Packer, r.SequenceNum)
}

func startActiveTest(c *conn) {
	exceed, done := make(chan struct{}), make(chan struct{})
	c.done = done
	c.exceed = exceed

	go func() {
		t := time.NewTicker(c.t)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if atomic.LoadInt32(&c.counter) >= c.n {
					c.server.ErrorLog.Printf("no smpp enquire link response returned from %v for %d times!",
						c.Conn.RemoteAddr(), c.n)
					exceed <- struct{}{}
					break
				}
				p := &pkg.SmppEnquireLinkReqPkt{}
				err := c.Conn.SendPkt(p, <-c.Conn.SequenceNum)
				if err != nil {
					c.server.ErrorLog.Printf("send smpp enquire link request to %v error: %v", c.Conn.RemoteAddr(), err)
				} else {
					atomic.AddInt32(&c.counter, 1)
				}
			}
		}
	}()
}

func (c *conn) serve() {
	defer func() {
		if err := recover(); err != nil {
			c.server.ErrorLog.Printf("panic serving %v: %v\n", c.Conn.RemoteAddr(), err)
		}
	}()

	defer c.close()

	startActiveTest(c)

	for {
		select {
		case <-c.exceed:
			return
		default:
		}

		r, err := c.readPacket()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				continue
			}
			break
		}

		_, err = c.server.Handler.ServeSmpp(r, r.Packet)
		if err1 := c.finishPacket(r); err1 != nil {
			break
		}

		if err != nil {
			break
		}
	}
}

func (srv *Server) newConn(rwc net.Conn) (c *conn, err error) {
	c = new(conn)
	c.server = srv
	c.readTimeout = c.server.ReadTimeout
	c.Conn = pkg.NewConnection(rwc, srv.Version)
	c.Conn.SetState(pkg.CONNECTION_CONNECTED)
	c.n = c.server.N
	c.t = c.server.T
	return c, nil
}

func (srv *Server) listenAndServe() error {
	if srv.Addr == "" {
		return ErrEmptyServerAddr
	}
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
}

func ListenAndServe(addr string, version uint8, t, readTimeout time.Duration, n int32, logWriter io.Writer, handlers ...Handler) error {
	if addr == "" {
		return ErrEmptyServerAddr
	}

	if handlers == nil {
		return ErrNoHandlers
	}

	var handler Handler
	handler = HandlerFunc(func(r *Response, p *Packet) (bool, error) {
		for _, h := range handlers {
			next, err := h.ServeSmpp(r, p)
			if err != nil || !next {
				return next, err
			}
		}
		return false, nil
	})

	if logWriter == nil {
		logWriter = os.Stderr
	}
	server := &Server{
		Addr:        addr,
		Handler:     handler,
		Version:     version,
		ReadTimeout: readTimeout,
		T:           t,
		N:           n,
		ErrorLog:    log.New(logWriter, "smpp server: ", log.LstdFlags)}
	return server.listenAndServe()
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(1 * time.Minute) // 1min
	return tc, nil
}
