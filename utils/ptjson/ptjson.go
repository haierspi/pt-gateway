package ptjson

import (
	jsoniter "github.com/json-iterator/go"
)

// IsValidJSON 验证JSON
func IsValidJSON(data []byte) bool {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	var j = make(map[string]interface{})
	err := json.Unmarshal(data, &j)
	if err != nil {
		return false
	}
	return true
}

// PrettyMarshal 美化JSON
func PrettyMarshal(src interface{}) ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.MarshalIndent(src, "", "  ")
}

// Marshal 序列化
func Marshal(src interface{}) ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(src)
}

// Unmarshal 反序列化
func Unmarshal(data []byte, v interface{}) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, v)
}
