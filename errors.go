package netti

import "errors"

var (
	// ErrProtocolNotSupported 当尝试使用不受支持的协议时发生
	ErrProtocolNotSupported = errors.New("not supported protocol on this platform")
	// ErrServerShutdown 在服务器关闭时发生
	ErrServerShutdown = errors.New("server is going to be shutdown")
	// ErrInvalidFixedLength 当输出数据具有无效的固定长度时发生
	ErrInvalidFixedLength = errors.New("invalid fixed length of bytes")
	// ErrUnexpectedEOF 当没有足够的数据可供编解码器读取时发生
	ErrUnexpectedEOF = errors.New("there is no enough data")
	// ErrDelimiterNotFound 当输入数据中没有此类分隔符时发生
	ErrDelimiterNotFound = errors.New("there is no such a delimiter")
	// ErrCRLFNotFound 当编码解码器找不到CRLF时发生
	ErrCRLFNotFound = errors.New("there is no CRLF")
	// ErrUnsupportedLength 当不支持的lengthFieldLength来自输入数据时发生
	ErrUnsupportedLength = errors.New("unsupported lengthFieldLength. (expected: 1, 2, 3, 4, or 8)")
	// ErrTooLessLength 当调整帧长度小于零时发生
	ErrTooLessLength = errors.New("adjusted frame length is less than zero")
)
