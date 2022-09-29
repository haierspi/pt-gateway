package rpc

// CommonReply 用于内部API，三个标准字段
type CommonReply struct {
	ErrorCode int64
	ErrorMsg  string
	Data      interface{}
}

// BodyReply 用于提供给三方调用，如支付回调，Body为内容，ContentType为文本类型：application/json，text/plain，text/xml等
type BodyReply struct {
	Body        []byte
	ContentType string
}
