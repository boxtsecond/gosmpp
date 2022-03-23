package pkg

import (
	"errors"
	"strconv"
)

type Status uint32

func (s *Status) Data() uint32 {
	return uint32(*s)
}

func (s *Status) Error() error {
	return errors.New(strconv.Itoa(int(*s)) + " : " + s.String())
}

func (s Status) String() string {

	var msg string
	switch s {
	case ESME_ROK:
		msg = "成功"
	case ESME_RINVMSGLEN:
		msg = "消息长度错"
	case ESME_RINVCMDLEN:
		msg = "命令长度错"
	case ESME_RINVCMDID:
		msg = "无效的命令 ID"
	case ESME_RINVBNDSTS:
		msg = "命令与 Bind 状态不一致"
	case ESME_RALYBND:
		msg = "ESME 已经绑定"
	case ESME_RINVPRTFLG:
		msg = "无效的优先标识"
	case ESME_RINVREGDLVFLG:
		msg = "无效状态报告标识"
	case ESME_RSYSERR:
		msg = "系统错"
	case ESME_RINVSRCADR:
		msg = "源地址无效"
	case ESME_RINVDSTADR:
		msg = "目标地址错"
	case ESME_RINVMSGID:
		msg = "消息 ID 错"
	case ESME_RBINDFAIL:
		msg = "绑定失败"
	case ESME_RINVPASWD:
		msg = "密码错误"
	case ESME_RINVSYSID:
		msg = "系统 ID 错误"
	case ESME_RCANCELFAIL:
		msg = "Cancel 消息 失败"
	case ESME_RREPLACEFAIL:
		msg = "Replace 消息失败"
	case ESME_RMSGQFUL:
		msg = "消息队列满"
	case ESME_RINVSERTYP:
		msg = "服务类型非法"
	case ESME_RINVNUMDESTS:
		msg = "目标号错误"
	case ESME_RINVDLNAME:
		msg = "名字分配表错误"
	case ESME_RINVDESTFLAG:
		msg = "目标标识错误"
	case ESME_RINVSUBREP:
		msg = "无效的 submit with replace 请求（如 sumit_sm 操作中 replace_if_present_flag 已设置）"
	case ESME_RINVESMCLASS:
		msg = "esm_class 字段数据非法"
	case ESME_RCNTSUBDL:
		msg = "无法提交至分配表"
	case ESME_RSUBMITFAIL:
		msg = "submit_sm 或 submit_muli 失败"
	case ESME_RINVSRCTON:
		msg = "无效的源地址 TON"
	case ESME_RINVSRCNPI:
		msg = "无效的源地址 NPI"
	case ESME_RINVDSTTON:
		msg = "无效的目标地址 TON"
	case ESME_RINVDSTNPI:
		msg = "无效的目标地址 NPI"
	case ESME_RINVSYSTYP:
		msg = "System_type 字段无效"
	case ESME_RINVREPFLAG:
		msg = "replace_if_present_flag 字段无效"
	case ESME_RINVNUMMSGS:
		msg = "消息序号无效"
	case ESME_RTHROTTLED:
		msg = "节流错（ESME 超出消息限制）"
	case ESME_RINVSCHED:
		msg = "无效的定时时间"
	case ESME_RINVEXPIRY:
		msg = "无效的超时时间"
	case ESME_RINVDFTMSGID:
		msg = "预定义消息无效或不存在"
	case ESME_RX_T_APPN:
		msg = "ESME 接收端暂时出错"
	case ESME_RX_P_APPN:
		msg = "ESME 接收端永久出错"
	case ESME_RX_R_APPN:
		msg = "ESME 接收端拒绝消息出错"
	case ESME_RQUERYFAIL:
		msg = "Query_sm 失败"
	case ESME_RINVOPTPARSTREAM:
		msg = "PDU 报体可选部分出错"
	case ESME_VOPTPARNOTALLWD:
		msg = "可选参数不允许"
	case ESME_RINVPARLEN:
		msg = "参数长度错"
	case ESME_RMISSINGOPTPARAM:
		msg = "需要的可选参数丢失"
	case ESME_RINVOPTPARAMVAL:
		msg = "无效的可选参数值"
	case ESME_RDELIVERYFAILURE:
		msg = "下发消息失败（用于data_sm_resp）"
	case ESME_RUNKNOWNERR:
		msg = "不明错误"

	default:
		msg = "Status Unknown: " + strconv.Itoa(int(s))
	}

	return msg
}

const (
	ESME_ROK Status = iota
	ESME_RINVMSGLEN
	ESME_RINVCMDLEN
	ESME_RINVCMDID
	ESME_RINVBNDSTS
	ESME_RALYBND
	ESME_RINVPRTFLG
	ESME_RINVREGDLVFLG
	ESME_RSYSERR
	_
	ESME_RINVSRCADR
	ESME_RINVDSTADR
	ESME_RINVMSGID
	ESME_RBINDFAIL
	ESME_RINVPASWD
	ESME_RINVSYSID
	_
	ESME_RCANCELFAIL
	_
	ESME_RREPLACEFAIL
	ESME_RMSGQFUL
	ESME_RINVSERTYP
)

const (
	ESME_RINVDESTFLAG Status = 0x00000040 + iota
	_
	ESME_RINVSUBREP
	ESME_RINVESMCLASS
	ESME_RCNTSUBDL
	ESME_RSUBMITFAIL
	_
	_
	ESME_RINVSRCTON
	ESME_RINVSRCNPI
	ESME_RINVDSTTON
	ESME_RINVDSTNPI
	_
	ESME_RINVSYSTYP
	ESME_RINVREPFLAG
	ESME_RINVNUMMSGS
	_
	_
	ESME_RTHROTTLED
	_
	_
	ESME_RINVSCHED
	ESME_RINVEXPIRY
	ESME_RINVDFTMSGID
	ESME_RX_T_APPN
	ESME_RX_P_APPN
	ESME_RX_R_APPN
	ESME_RQUERYFAIL
)

const (
	ESME_RINVOPTPARSTREAM Status = 0x000000C0 + iota
	ESME_VOPTPARNOTALLWD
	ESME_RINVPARLEN
	ESME_RMISSINGOPTPARAM
	ESME_RINVOPTPARAMVAL
)

const (
	ESME_RINVNUMDESTS, ESME_RINVDLNAME      Status = 0x00000033, 0x00000034
	ESME_RDELIVERYFAILURE, ESME_RUNKNOWNERR Status = 0x000000FE, 0x000000FF
)
