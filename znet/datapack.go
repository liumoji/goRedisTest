package znet

import (
	"strconv"
	"strings"
	"zinx/utils"
	"zinx/ziface"
)

//封包拆包类实例，暂时不需要成员
// type DataPack struct {}

// //封包拆包实例初始化方法
// func NewDataPack() *DataPack {
// 	return &DataPack{}
// }

// //获取包头长度方法
// func(dp *DataPack) GetHeadLen() uint32 {
// 	//Id uint32(4字节) +  DataLen uint32(4字节)
// 	return 8
// }
//封包方法(压缩数据)
// func(dp *DataPack) Pack(msg ziface.IMessage)([]byte, error) {
// 	//创建一个存放bytes字节的缓冲
// 	dataBuff := bytes.NewBuffer([]byte{})

// 	//写dataLen
// 	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
// 		return nil, err
// 	}

// 	//写msgID
// 	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
// 		return nil, err
// 	}

// 	//写data数据
// 	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
// 		return nil ,err
// 	}

// 	return dataBuff.Bytes(), nil
// }
//拆包方法(解压数据)
// func(dp *DataPack) Unpack(binaryData []byte)(ziface.IMessage, error) {
// 	//创建一个从输入二进制数据的ioReader
// 	dataBuff := bytes.NewReader(binaryData)

// 	//只解压head的信息，得到dataLen和msgID
// 	msg := &Message{}

// 	//读dataLen
// 	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
// 		return nil, err
// 	}

// 	//读msgID
// 	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
// 		return nil, err
// 	}

// 	//判断dataLen的长度是否超出我们允许的最大包长度
// 	if (utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize) {
// 		return nil, errors.New("Too large msg data recieved")
// 	}

// 	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
// 	return msg, nil
// }

//新增resp协议解析=======================================================
//封包拆包类实例
type RespDataPack struct {
	arrlen int      //数组长度,为-1时表示解析未开始或当前解析已完成
	arridx int      //当前解析完成的数组项数
	offset int      //当前解析的input偏移
	input  string   //输入字符串,也就是要解析的字符串
	cmd    []string //解析出的命令
}

const (
	RESP_OK          = 0  //成功解析
	RESP_ERR_PENDING = -1 //数据不完整, 等待
	RESP_ERR_INVALID = -2 //数据格式非法
	RESP_ERR_NOINPUT = -3 //无待解析数据
)

const (
	CRLF                 = "\r\n"
	PREFIX_SIMPLE_STRING = "+"
	PREFIX_ERROR         = '-'
	PREFIX_INT           = ':'
	PREFIX_BULK_STRING   = '$'
	PREFIX_ARRAY         = '*'
)

//封包拆包实例初始化方法
func NewRespDataPack() *RespDataPack {
	return &RespDataPack{}
}

func (rdp *RespDataPack) RespInit() {
	rdp.arrlen = -1
	rdp.arridx = 0
	rdp.offset = 0
	rdp.input = ""
	rdp.cmd = make([]string, 0, 5)
}

func (rdp *RespDataPack) roundover() {
	rdp.arrlen = -1
	rdp.arridx = 0
	rdp.offset = 0
	rdp.cmd = rdp.cmd[0:0] //清空切片
}

//解析resp格式字符串
func (rdp *RespDataPack) RespUnpack(input string) int {
	if len(input) != 0 {
		//上次信息未处理完整，拼接后继续解析
		rdp.input += input
	}
	if len(rdp.input) == 0 {
		return RESP_ERR_NOINPUT
	}
	if rdp.arrlen == -1 {
		//数据尚未解析过
		rdp.roundover()
		if rdp.input[rdp.offset] != PREFIX_ARRAY {
			return RESP_ERR_INVALID
		}
		crlfIdx := strings.Index(rdp.input, CRLF)
		if crlfIdx == -1 {
			//TODO:判断到底是还没收到足够的数据还是数据格式有问题
			return RESP_ERR_PENDING
		}

		arrLenStr := rdp.input[rdp.offset+1 : crlfIdx]
		//判断是非0自然数
		isNaturalNumber, _ := utils.IsNaturalNumber(arrLenStr)
		if isNaturalNumber {
			rdp.arrlen, _ = strconv.Atoi(arrLenStr) //获取到数组长度
			rdp.offset += (crlfIdx + 2)
			rdp.input = rdp.input[rdp.offset:len(rdp.input)]
			rdp.offset = 0
			if rdp.offset+4 >= len(rdp.input) {
				return RESP_ERR_PENDING
			}
			return rdp.RespUnpack("")
		} else {
			return RESP_ERR_INVALID
		}
	} else {
		//数组长度已经解析
		if rdp.input[rdp.offset] != PREFIX_BULK_STRING {
			return RESP_ERR_INVALID
		}
		//预期解析不出字符串长度
		if rdp.offset+4 >= len(rdp.input) {
			return RESP_ERR_PENDING
		}

		crlfIdx := strings.Index(rdp.input, CRLF)
		if crlfIdx == -1 {
			//TODO:判断到底是还没收到足够的数据还是数据格式有问题
			return RESP_ERR_PENDING
		}
		arrLenStr := rdp.input[rdp.offset+1 : crlfIdx]
		isNaturalNumber, _ := utils.IsNaturalNumber(arrLenStr)
		if isNaturalNumber || arrLenStr == "0" {
			strlen, _ := strconv.Atoi(arrLenStr) //获取到字符串的长度
			if len(rdp.input) >= crlfIdx+strlen+4 {
				if rdp.input[crlfIdx+strlen+2:crlfIdx+strlen+4] == CRLF {
					rdp.cmd = append(rdp.cmd,
						rdp.input[crlfIdx+2:crlfIdx+2+strlen])
					rdp.input = rdp.input[crlfIdx+strlen+4 : len(rdp.input)]
					rdp.offset = 0
					rdp.arridx++
					if rdp.arridx == rdp.arrlen {
						rdp.arrlen = -1
						return RESP_OK
					} else {
						return rdp.RespUnpack("")
					}
				} else {
					return RESP_ERR_INVALID
				}
			} else {
				return RESP_ERR_PENDING
			}
		} else {
			return RESP_ERR_INVALID
		}
	}
}

//封包方法(压缩数据)
func (rdp *RespDataPack) RespPack(msg ziface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	//dataBuff := bytes.NewBuffer([]byte{})

	repMsg := createSimpleString(msg.GetData()[0])
	repBuf := []byte(repMsg)

	// //写data数据
	// if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
	// 	return nil ,err
	// }

	//return dataBuff.Bytes(), nil
	return repBuf, nil
}

func createSimpleString(input string) string {
	return PREFIX_SIMPLE_STRING + input + CRLF
}
