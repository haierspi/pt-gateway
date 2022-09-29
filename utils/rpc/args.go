package rpc

// BodyArgs 用于提供给三方调用，如支付回调，Body为内容
type BodyArgs struct {
	Body string
}

// EmptyArgs 空参数
type EmptyArgs struct {
}

// TokenAndDeviceIDArgs 只有Token和DeviceID的参数
type TokenAndDeviceIDArgs struct {
	Token    string
	DeviceID string
}
