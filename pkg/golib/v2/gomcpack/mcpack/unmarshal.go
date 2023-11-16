package mcpack

import (
	"fmt"
	"strings"
)

var (
	InvalidPackError = fmt.Errorf("the pack is invalid")
)

type Tag string

const (
	TagPack = Tag("PCK")
	TagItem = Tag("ITM")
)

// V2Map represents common pack object
type V2Map map[string]interface{}

// Unmarshal V2Map from byte buffer
func Decode(buf []byte) (V2Map, error) {
	if len(buf) < 4 {
		return nil, InvalidPackError
	}
	switch Tag(buf[0:4]) {
	case TagPack:
		// this is v1 pack, does not support for now
		return nil, InvalidPackError
	default:
		// consider everthing else as v2
	}

	idx := &v2PackIdx{}
	err := idx.buildItemsFromBuffer(buf)
	if err != nil {
		return nil, err
	}

	res, err := bind(idx.root)
	if err != nil {
		return nil, err
	}
	if mapres, ok := res.(V2Map); ok {
		return mapres, nil
	}

	return nil, fmt.Errorf("invalid root data type %v", res)
}

// Get the specified key from pack
func (m V2Map) Get(key string) interface{} {
	kArr := strings.Split(key, ".")
	res := interface{}(m)
	for _, k := range kArr {
		if resMap, ok := res.(V2Map); ok {
			if val, exists := resMap[k]; exists {
				res = val
				continue
			}
		}
		res = nil
		break
	}

	return res
}

// GetString of the specified key
func (m V2Map) GetString(key, def string) string {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(string); ok {
		return actualRes
	}
	return def
}

// GetBool of the specified key
func (m V2Map) GetBool(key string, def bool) bool {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(bool); ok {
		return actualRes
	}
	return def
}

// GetInt8 of the specified key
func (m V2Map) GetInt8(key string, def int8) int8 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(int8); ok {
		return actualRes
	}
	return def
}

// GetInt16 of the specified key
func (m V2Map) GetInt16(key string, def int16) int16 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(int16); ok {
		return actualRes
	}
	return def
}

// GetInt32 of the specified key
func (m V2Map) GetInt32(key string, def int32) int32 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(int32); ok {
		return actualRes
	}
	return def
}

// GetInt64 of the specified key
func (m V2Map) GetInt64(key string, def int64) int64 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(int64); ok {
		return actualRes
	}
	return def
}

// GetUint8 of the specified key
func (m V2Map) GetUint8(key string, def uint8) uint8 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(uint8); ok {
		return actualRes
	}
	return def
}

// GetUint16 of the specified key
func (m V2Map) GetUint16(key string, def uint16) uint16 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(uint16); ok {
		return actualRes
	}
	return def
}

// GetUint32 of the specified key
func (m V2Map) GetUint32(key string, def uint32) uint32 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(uint32); ok {
		return actualRes
	}
	return def
}

// GetUint64 of the specified key
func (m V2Map) GetUint64(key string, def uint64) uint64 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(uint64); ok {
		return actualRes
	}
	return def
}

// GetFloat32 of the specified key
func (m V2Map) GetFloat32(key string, def float32) float32 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(float32); ok {
		return actualRes
	}
	return def
}

// GetFloat64 of the specified key
func (m V2Map) GetFloat64(key string, def float64) float64 {
	res := m.Get(key)
	if res == nil {
		return def
	}
	if actualRes, ok := res.(float64); ok {
		return actualRes
	}
	return def
}
