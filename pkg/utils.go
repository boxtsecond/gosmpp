package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func GenTimestamp() uint32 {
	s := time.Now().Format("0102150405")
	i, _ := strconv.Atoi(s)
	return uint32(i)
}

func GenTimestampYYStr(t int64) string {
	return time.Unix(t, 0).Format("0601021504")
}

func GenNowTimeYYYYStr() string {
	s := time.Now().Format("20060102150405")
	return s
}

func GenNowTimeYYStr() string {
	return time.Unix(time.Now().Unix(), 0).Format("0601021504")
}

/*
	生成算法:
	时间秒+序列号(顺序增加，步长为1，循环使用)
*/
func GenMsgID(sequenceNum uint32) string {
	now := time.Now()
	sec := now.Second()
	seqStr := fmt.Sprintf("%08d", sequenceNum)
	if len(seqStr) > 8 {
		seqStr = seqStr[len(seqStr)-8:]
	}
	return fmt.Sprintf("%02d%s", sec, seqStr)
}

func Utf8ToUcs2(in string) (string, error) {
	if !utf8.ValidString(in) {
		return "", errors.New("invalid utf8 runes")
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()) //UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Ucs2ToUtf8(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()) //UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Utf8ToGB18030(in string) (string, error) {
	if !utf8.ValidString(in) {
		return "", errors.New("invalid utf8 runes")
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, simplifiedchinese.GB18030.NewEncoder())
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func GB18030ToUtf8(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, simplifiedchinese.GB18030.NewDecoder())
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func GetUtf8Content(msgFmt uint8, msgContent string) (string, error) {
	switch msgFmt {
	case ASCII:
	case BINARY:
	case UCS2:
		return Ucs2ToUtf8(msgContent)
	case GB18030:
		return GB18030ToUtf8(msgContent)
	default:
	}

	return "", errors.New("invalid msg fmt")
}

var TpUdhiSeq byte = 0x00

func SplitLongSms(content string) [][]byte {
	smsLength := 140
	smsHeaderLength := 6
	smsBodyLen := smsLength - smsHeaderLength
	contentBytes := []byte(content)
	var chunks [][]byte
	num := 1
	if (len(content)) > 140 {
		num = int(math.Ceil(float64(len(content)) / float64(smsBodyLen)))
	}
	if num == 1 {
		chunks = append(chunks, contentBytes)
		return chunks
	}
	tpUdhiHeader := []byte{0x05, 0x00, 0x03, TpUdhiSeq, byte(num)}
	TpUdhiSeq++

	for i := 0; i < num; i++ {
		chunk := tpUdhiHeader
		chunk = append(chunk, byte(i+1))
		bodyLen := smsLength - smsHeaderLength
		offset := i * bodyLen
		max := offset + bodyLen
		if max > len(content) {
			max = len(content)
		}

		chunk = append(chunk, contentBytes[offset:max]...)
		chunks = append(chunks, chunk)
	}
	return chunks
}

func GetMsgPkgs(pkg *SmppSubmitReqPkt) ([]*SmppSubmitReqPkt, error) {
	packets := make([]*SmppSubmitReqPkt, 0)
	content, err := Utf8ToUcs2(pkg.ShortMessage)
	if err != nil {
		return packets, err
	}

	chunks := SplitLongSms(content)
	var esmClass uint8
	if len(chunks) > 1 {
		esmClass = SM_UDH_GSM
	}

	for _, chunk := range chunks {
		p := &SmppSubmitReqPkt{
			ServiceType:        pkg.ServiceType,
			SourceAddrTON:      pkg.SourceAddrTON,
			SourceAddrNPI:      pkg.SourceAddrNPI,
			SourceAddr:         pkg.SourceAddr,
			DestAddrTON:        pkg.DestAddrTON,
			DestAddrNPI:        pkg.DestAddrNPI,
			DestinationAddr:    pkg.DestinationAddr, // phone
			EsmClass:           esmClass,
			PriorityFlag:       NORMAL_PRIORITY,
			RegisteredDelivery: NEED_REPORT,
			DataCoding:         UCS2,
			SmLength:           uint8(len(chunk)),
			ShortMessage:       string(chunk),
		}
		packets = append(packets, p)
	}
	return packets, nil
}
