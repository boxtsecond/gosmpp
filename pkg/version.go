package pkg

const (
	VERSION uint8 = 0x34
)

// EsmClass
const (
	SM_UDH_GSM uint8 = 0x40
	SM_DELIVER uint8 = 4
)

// DataCoding
// 短消息内容体的编码格式
const (
	ASCII   = 0  // ASCII编码
	LATIN1  = 3  // Latin 1
	BINARY  = 4  // 二进制短消息
	UCS2    = 8  // UCS2编码
	GB18030 = 15 // GB18030编码
)

const (
	NOT_REPORT = 0 // 不是状态报告
	IS_REPORT  = 1 // 是状态报告
)

// 是否要求返回状态报告
const (
	NO_NEED_REPORT     = 0
	NEED_REPORT        = 1
	ONLY_FAILED_REPORT = 2
)

// 短消息发送优先级
const (
	NORMAL_PRIORITY = iota
	HIGH_PRIORITY
	HIGHER_PRIORITY
	HIGHEST_PRIORITY
)
