package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	SMPP_PACKET_MAX uint32 = 2477
	SMPP_PACKET_MIN uint32 = 16
)

type Packer interface {
	Pack(seqId uint32) ([]byte, error)
	Unpack(data []byte) error
	String() string
}

type pkgWriter struct {
	b   *bytes.Buffer
	err *OpError
}

func newPkgWriter(initSize uint32) *pkgWriter {
	buf := make([]byte, 0, initSize)
	return &pkgWriter{
		b: bytes.NewBuffer(buf),
	}
}

func (w *pkgWriter) Bytes() ([]byte, error) {
	if w.err != nil {
		return nil, w.err.err
	}
	l := w.b.Len()
	return (w.b.Bytes())[:l], nil
}

func (w *pkgWriter) WriteByte(b byte) {
	if w.err != nil {
		return
	}

	err := w.b.WriteByte(b)
	if err != nil {
		w.err = NewOpError(err,
			fmt.Sprintf("pkgWriter.WriteByte writes: %x", b))
		return
	}
}

func (w *pkgWriter) WriteBytes(b []byte) {
	if w.err != nil {
		return
	}

	_, err := w.b.Write(b)
	if err != nil {
		w.err = NewOpError(err,
			fmt.Sprintf("pkgWriter.WriteBytes writes: %x", b))
		return
	}
}

func (w *pkgWriter) WriteFixedSizeString(s string, size int) {
	if w.err != nil {
		return
	}

	l1 := len(s)
	l2 := l1
	if l2 > 10 {
		l2 = 10
	}

	if l1 > size {
		w.err = NewOpError(ErrMethodParamsInvalid,
			fmt.Sprintf("pkgWriter.WriteFixedSizeString writes: %s", s[0:l2]))
		return
	}

	w.WriteString(strings.Join([]string{s, string(make([]byte, size-l1))}, ""))
}

func (w *pkgWriter) WriteString(s string) {
	if w.err != nil {
		return
	}

	l1 := len(s)
	l2 := l1
	if l2 > 10 {
		l2 = 10
	}

	n, err := w.b.WriteString(s)
	if err != nil {
		w.err = NewOpError(err,
			fmt.Sprintf("pkgWriter.WriteString writes: %s", s[0:l2]))
		return
	}

	if n != l1 {
		w.err = NewOpError(fmt.Errorf("WriteString writes %d bytes, not equal to %d we expected", n, l1),
			fmt.Sprintf("pkgWriter.WriteString writes: %s", s[0:l2]))
		return
	}
}

func (w *pkgWriter) WriteInt(order binary.ByteOrder, data interface{}) {
	if w.err != nil {
		return
	}

	err := binary.Write(w.b, order, data)
	if err != nil {
		w.err = NewOpError(err,
			fmt.Sprintf("pkgWriter.WriteInt writes: %#v", data))
		return
	}
}

func (w *pkgWriter) WriteHeader(header Header) {
	w.WriteInt(binary.BigEndian, header.CommandLength)
	w.WriteInt(binary.BigEndian, header.CommandID)
	w.WriteInt(binary.BigEndian, header.CommandStatus)
	w.WriteInt(binary.BigEndian, header.SequenceNum)
}

const maxCStringSize = 160

type pkgReader struct {
	rb   *bytes.Buffer
	err  *OpError
	cbuf [maxCStringSize]byte
}

func newPkgReader(data []byte) *pkgReader {
	return &pkgReader{
		rb: bytes.NewBuffer(data),
	}
}

func (r *pkgReader) ReadByte() byte {
	if r.err != nil {
		return 0
	}

	b, err := r.rb.ReadByte()
	if err != nil {
		r.err = NewOpError(err,
			"pkgReader.ReadByte")
		return 0
	}
	return b
}

func (r *pkgReader) ReadInt(order binary.ByteOrder, data interface{}) {
	if r.err != nil {
		return
	}

	err := binary.Read(r.rb, order, data)
	if err != nil {
		r.err = NewOpError(err,
			"pkgReader.ReadInt")
		return
	}
}

func (r *pkgReader) ReadBytes(s []byte) {
	if r.err != nil {
		return
	}

	n, err := r.rb.Read(s)
	if err != nil {
		r.err = NewOpError(err,
			"pkgReader.ReadBytes")
		return
	}

	if n != len(s) {
		r.err = NewOpError(fmt.Errorf("ReadBytes reads %d bytes, not equal to %d we expected", n, len(s)),
			"pkgReader.ReadBytes")
		return
	}
}

func (r *pkgReader) ReadCString(length int) []byte {
	if r.err != nil {
		return nil
	}

	var tmp = r.cbuf[:length]
	n, err := r.rb.Read(tmp)
	if err != nil {
		r.err = NewOpError(err,
			"pkgReader.ReadCString")
		return nil
	}

	if n != length {
		r.err = NewOpError(fmt.Errorf("ReadCString reads %d bytes, not equal to %d we expected", n, length),
			"pkgWriter.ReadCString")
		return nil
	}

	i := bytes.IndexByte(tmp, 0)
	if i == -1 {
		return tmp
	} else {
		return tmp[:i]
	}
}

func (r *pkgReader) ReadOCString(maxLength int) []byte {
	if r.err != nil {
		return nil
	}

	line, err := r.rb.ReadBytes(COctetStringNULL)
	if err != nil {
		r.err = NewOpError(err,
			"pkgReader.ReadOCString")
		return nil
	}

	if len(line) == 0 {
		return nil
	}

	if len(line) > maxLength {
		r.err = NewOpError(fmt.Errorf("ReadOCString reads %d bytes, greater than %d we expected", len(line), maxLength),
			"pkgWriter.ReadOCString")
		return nil
	}
	return line[:len(line)-1]
}

func (r *pkgReader) ReadOCStringBySpace() []byte {
	if r.err != nil {
		return nil
	}

	line, err := r.rb.ReadBytes(byte(' '))
	if err != nil {
		r.err = NewOpError(err,
			"pkgReader.ReadOCString")
		return nil
	}

	if len(line) == 0 {
		return nil
	}

	return line[:len(line)-1]
}

func (r *pkgReader) Error() error {
	if r.err != nil {
		return r.err
	}
	return nil
}
