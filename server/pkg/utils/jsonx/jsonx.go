package jsonx

import (
	"encoding/json"
	"mayfly-go/pkg/logx"
	"mayfly-go/pkg/utils/anyx"
	"mayfly-go/pkg/utils/collx"

	"github.com/tidwall/gjson"
)

// To json字节数组转指定类型
func To[T any](jsonVal []byte) (*T, error) {
	var v T
	if err := json.Unmarshal(jsonVal, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ToMap json字节数组转map
func ToMap(jsonVal []byte) (collx.M, error) {
	var res map[string]any
	err := json.Unmarshal(jsonVal, &res)
	return res, err
}

// ToByStr json字符串转指定类型
func ToByStr[T any](jsonStr string) (*T, error) {
	return To[T]([]byte(jsonStr))
}

// ToMapByStr json字符串转map
func ToMapByStr(jsonStr string) (collx.M, error) {
	if jsonStr == "" {
		return map[string]any{}, nil
	}
	return ToMap([]byte(jsonStr))
}

// 转换为json字符串
func ToStr(val any) string {
	if anyx.IsBlank(val) {
		return ""
	}
	if strBytes, err := json.Marshal(val); err != nil {
		logx.ErrorTrace("toJsonStr error: ", err)
		return ""
	} else {
		return string(strBytes)
	}
}

// 根据json字节数组获取对应字段路径的string类型值
//
//   - fieldPath字段路径。如user.username等
func GetStringByBytes(bytes []byte, fieldPath string) (string, error) {
	return gjson.GetBytes(bytes, fieldPath).String(), nil
}

// 根据json字符串获取对应字段路径的string类型值
//
//   - fieldPath字段路径。如user.username等
func GetString(jsonStr string, fieldPath string) (string, error) {
	return gjson.Get(jsonStr, fieldPath).String(), nil
}

// 根据json字节数组获取对应字段路径的int类型值
//
//   - fieldPath字段路径。如user.age等
func GetIntByBytes(bytes []byte, fieldPath string) (int64, error) {
	return gjson.GetBytes(bytes, fieldPath).Int(), nil
}

// 根据json字符串获取对应字段路径的int类型值
//
//   - fieldPath字段路径。如user.age等
func GetInt(jsonStr string, fieldPath string) (int64, error) {
	return gjson.Get(jsonStr, fieldPath).Int(), nil
}

// 根据json字节数组获取对应字段路径的bool类型值
//
//   - fieldPath字段路径。如user.isDeleted等
func GetBoolByBytes(bytes []byte, fieldPath string) (bool, error) {
	return gjson.GetBytes(bytes, fieldPath).Bool(), nil
}

// 根据json字符串获取对应字段路径的bool类型值
//
//   - fieldPath字段路径。如user.isDeleted等
func GetBool(jsonStr string, fieldPath string) (bool, error) {
	return GetBoolByBytes([]byte(jsonStr), fieldPath)
}
