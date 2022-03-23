package pkg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var ErrLength = errors.New("Options: error length")

type Tag uint16

// 可选参数标签定义  Option Tag
const (
	TAG_DestAddrSubunit Tag = 0x0005 + iota
	TAG_DestNetworkType
	TAG_DestBearerType
	TAG_DestTelematicsId
)

const (
	TAG_SourceAddrSubunit Tag = 0x000D + iota
	TAG_SourceNetworkType
	TAG_SourceBearerType
	TAG_SourceTelematicsId
)

const (
	TAG_PrivacyIndicator Tag = 0x0201 + iota
	TAG_SourceSubaddress
	TAG_DestSubaddress
	TAG_UserMessageReference
	TAG_UserResponseCode
)

const (
	TAG_SourcePort Tag = 0x020A + iota
	TAG_DestinationPort
	TAG_SarMsgRefNum
	TAG_LanguageIndicator
	TAG_SarTotalSegments
	TAG_SarSegmentSeqnum
	TAG_SCInterfaceVersion
)

const (
	TAG_DpfResult Tag = 0x0420 + iota
	TAG_SetDpf
	TAG_MsAvailabilityStatus
	TAG_NetworkErrorCode
	TAG_MessagePayload
	TAG_DeliveryFailureReason
	TAG_MoreMessagesToSend
	TAG_MessageState
)

const (
	TAG_QosTimeToLive            Tag = 0x0017
	TAG_PayloadType              Tag = 0x0019
	TAG_AdditionalStatusInfoText Tag = 0x001D
	TAG_ReceiptedMessageId       Tag = 0x001E
	TAG_MsMsgWaitFacilities      Tag = 0x0030
	TAG_CallbackNumPresIndt      Tag = 0x0302
	TAG_CallbackNumAtag          Tag = 0x0303
	TAG_NumberOfMessages         Tag = 0x0304
	TAG_CallbackNum              Tag = 0x0381
	TAG_UssdServiceOps           Tag = 0x0501
	TAG_DisplayTime              Tag = 0x1201
	TAG_SmsSignal                Tag = 0x1203
	TAG_MsValidity               Tag = 0x1204
	TAG_AlertOnMessageDelivery   Tag = 0x130C
	TAG_ItsReplyType             Tag = 0x1380
	TAG_ItsSessionInfo           Tag = 0x1383
)

var TagName = map[Tag]string{
	TAG_DestAddrSubunit:          "TAG_DestAddrSubunit",
	TAG_DestNetworkType:          "TAG_DestNetworkType",
	TAG_DestBearerType:           "TAG_DestBearerType",
	TAG_DestTelematicsId:         "TAG_DestTelematicsId",
	TAG_SourceAddrSubunit:        "TAG_SourceAddrSubunit",
	TAG_SourceNetworkType:        "TAG_SourceNetworkType",
	TAG_SourceBearerType:         "TAG_SourceBearerType",
	TAG_SourceTelematicsId:       "TAG_SourceTelematicsId",
	TAG_PrivacyIndicator:         "TAG_PrivacyIndicator",
	TAG_SourceSubaddress:         "TAG_SourceSubaddress",
	TAG_DestSubaddress:           "TAG_DestSubaddress",
	TAG_UserMessageReference:     "TAG_UserMessageReference",
	TAG_UserResponseCode:         "TAG_UserResponseCode",
	TAG_SourcePort:               "TAG_SourcePort",
	TAG_DestinationPort:          "TAG_DestinationPort",
	TAG_SarMsgRefNum:             "TAG_SarMsgRefNum",
	TAG_LanguageIndicator:        "TAG_LanguageIndicator",
	TAG_SarTotalSegments:         "TAG_SarTotalSegments",
	TAG_SarSegmentSeqnum:         "TAG_SarSegmentSeqnum",
	TAG_SCInterfaceVersion:       "TAG_SCInterfaceVersion",
	TAG_DpfResult:                "TAG_DpfResult",
	TAG_SetDpf:                   "TAG_SetDpf",
	TAG_MsAvailabilityStatus:     "TAG_MsAvailabilityStatus",
	TAG_NetworkErrorCode:         "TAG_NetworkErrorCode",
	TAG_MessagePayload:           "TAG_MessagePayload",
	TAG_DeliveryFailureReason:    "TAG_DeliveryFailureReason",
	TAG_MoreMessagesToSend:       "TAG_MoreMessagesToSend",
	TAG_MessageState:             "TAG_MessageState",
	TAG_QosTimeToLive:            "TAG_QosTimeToLive",
	TAG_PayloadType:              "TAG_PayloadType",
	TAG_AdditionalStatusInfoText: "TAG_AdditionalStatusInfoText",
	TAG_ReceiptedMessageId:       "TAG_ReceiptedMessageId",
	TAG_MsMsgWaitFacilities:      "TAG_MsMsgWaitFacilities",
	TAG_CallbackNumPresIndt:      "TAG_CallbackNumPresIndt",
	TAG_CallbackNumAtag:          "TAG_CallbackNumAtag",
	TAG_NumberOfMessages:         "TAG_NumberOfMessages",
	TAG_CallbackNum:              "TAG_CallbackNum",
	TAG_UssdServiceOps:           "TAG_UssdServiceOps",
	TAG_DisplayTime:              "TAG_DisplayTime",
	TAG_SmsSignal:                "TAG_SmsSignal",
	TAG_MsValidity:               "TAG_MsValidity",
	TAG_AlertOnMessageDelivery:   "TAG_AlertOnMessageDelivery",
	TAG_ItsReplyType:             "TAG_ItsReplyType",
	TAG_ItsSessionInfo:           "TAG_ItsSessionInfo",
}

// 可选参数 map
type Options map[Tag]*TLV

// 返回可选字段部分的长度
func (o Options) Len() int {
	length := 0

	for _, v := range o {
		length += 2 + 2 + int(v.Length)
	}

	return length
}

func (o Options) String() string {
	var b bytes.Buffer

	for _, v := range o {
		fmt.Fprintln(&b, "--- Options ---")
		fmt.Fprintln(&b, "Tag: ", v.Tag)
		fmt.Fprintln(&b, "Length: ", v.Length)
		fmt.Fprintln(&b, "Value: ", v.Value)
	}
	return b.String()
}

func ParseOptions(rawData []byte) (Options, error) {
	var (
		p      = 0
		ops    = make(Options)
		length = len(rawData)
	)

	for p < length {
		if length-p < 2+2 { // less than Tag len + Length len
			return nil, ErrLength
		}

		tag := binary.BigEndian.Uint16(rawData[p:])
		p += 2

		vlen := binary.BigEndian.Uint16(rawData[p:])
		p += 2

		if length-p < int(vlen) { // remaining not enough
			return nil, ErrLength
		}

		value := rawData[p : p+int(vlen)]
		p += int(vlen)

		ops[Tag(tag)] = NewTLV(Tag(tag), value)
	}

	return ops, nil
}
