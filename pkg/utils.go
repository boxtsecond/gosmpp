package pkg

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
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

func GenNowTimeYYYYStr() string {
	s := time.Now().Format("20060102150405")
	return s
}

func GenNowTimeYYStr() string {
	return time.Unix(time.Now().Unix(), 0).Format("0601021504")
}

// 生成客户端认证码
// 其值通过单向MD5 hash计算得出，表示如下：
// AuthenticatorClient =MD5（ClientID+7 字节的二进制0（0x00） + Shared secret+Timestamp）
// Shared secret 由服务器端与客户端事先商定，最长15字节。
// 此处Timestamp格式为：MMDDHHMMSS（月日时分秒），经TimeStamp字段值转换成字符串，转换后右对齐，左补0x30得到。
// 例如3月1日0时0分0秒，TimeStamp字段值为0x11F0E540，此处为0301000000。
func GenAuthenticatorClient(clientId, secret string, timestamp uint32) ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.WriteString(clientId)
	buf.Write([]byte{0, 0, 0, 0, 0, 0, 0})
	buf.WriteString(secret)
	buf.WriteString(fmt.Sprintf("%010d", timestamp))

	h := md5.New()
	_, err := h.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// 生成服务端认证码
// Login服务器端返回给客户端的认证码，当客户端认证出错时，此项为空。
// 其值通过单向MD5 hash计算得出，表示如下：
// AuthenticatorServer =MD5（Status+AuthenticatorClient + Shared secret）
// Shared secret 由服务器端与客户端事先商定,最长15字节AuthenticatorClient为客户端发送给服务器端的Login中的值。参见6.2.2节。
func GenAuthenticatorServer(status Status, secret, AuthenticatorClient string) ([]byte, error) {
	buf := new(bytes.Buffer)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, uint32(status))
	buf.Write(b)
	buf.WriteString(AuthenticatorClient)
	buf.WriteString(secret)

	h := md5.New()
	_, err := h.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

/*
	生成算法(64位整数):
	1. bit64 ~ bit39. 时间(格式为MMDDHHMMSS)
		a. bit64~bit61  月份的二进制表示
		b. bit60~bit56  日的二进制表示
		c. bit55~bit51  小时的二进制表示
		d. bit50~bit45  分钟的二进制表示
		e. bit44~bit39  秒的二进制表示
	2. bit38~bit17. 短信网关代码(转换为整数填充)
	3. bit16~bit1. 序列号(顺序增加，步长为1，循环使用)
	各部分若不能填满，左补零，右对齐
*/
func GenMsgID(spId string, sequenceNum uint32) (string, error) {
	now := time.Now()
	month, _ := strconv.ParseInt(fmt.Sprintf("%d", now.Month()), 10, 32)
	day := now.Day()
	hour := now.Hour()
	min := now.Minute()
	sec := now.Second()
	spIdInt, _ := strconv.ParseInt(spId, 10, 32)
	binaryStr := fmt.Sprintf("%04b%05b%05b%06b%06b%022b%016b", month, day, hour, min, sec, spIdInt, sequenceNum)
	msgId, err := strconv.ParseUint(binaryStr, 2, 64)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(msgId)), nil
}

func UnpackMsgId(msgId string) string {
	spId, _ := strconv.ParseUint(msgId[:6], 16, 24)
	month, _ := strconv.ParseUint(msgId[6:8], 10, 8)
	day, _ := strconv.ParseUint(msgId[8:10], 10, 8)
	hour, _ := strconv.ParseUint(msgId[10:12], 10, 8)
	min, _ := strconv.ParseUint(msgId[12:14], 10, 8)
	seqNum, _ := strconv.ParseUint(msgId[14:], 16, 24)
	return fmt.Sprintf("spId: %s, month: %d, day: %d, hour: %d, min: %d, seqNum: %d, ", NewOctetString(strconv.Itoa(int(spId))).FixedString(6), month, day, hour, min, seqNum)
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
