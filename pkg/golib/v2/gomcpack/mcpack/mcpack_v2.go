package mcpack

import (
	"encoding/binary"
	"fmt"
	"math"
)

type (
	V1PackItemType int32
	V2PackType     byte
)

const (
	V1PackBad    = V1PackItemType(0x00) /**< 非法       */
	V1PackPack   = V1PackItemType(0x01) /**< pack       */
	V1PackObject = V1PackItemType(0x02) /**< object       */
	V1PackArray  = V1PackItemType(0x04) /**< array       */

	V1ItemUnknown = V1PackItemType(0x05) /**< 未知类型，可能使用了版本更高的协议 */
	V1ItemBad     = V1PackItemType(0x00) /**< 非法       */
	V1ItemBinary  = V1PackItemType(0x10) /**< 二进制       */
	V1ItemText    = V1PackItemType(0x20) /**< 文本       */

	V1ItemSigned   = V1PackItemType(0x11) /**< 有符号       */
	V1ItemUnsigned = V1PackItemType(0x12) /**< 无符号       */
	V1Item32Bits   = V1PackItemType(0x14) /**< 32位       */
	V1Item64Bits   = V1PackItemType(0x18) /**< 64位       */

	V1ItemBool   = V1PackItemType(0x30) /**< BOOL类型   */
	V1ItemNull   = V1PackItemType(0x40) /**< NULL值     */
	V1ItemFloat  = V1PackItemType(0x50) /**< 4字节浮点  */
	V1ItemDouble = V1PackItemType(0x51) /**< 8字节浮点　*/

	V1ItemInt32  = V1PackItemType(V1ItemSigned | V1Item32Bits)   /**< int32       */
	V1ItemUint32 = V1PackItemType(V1ItemUnsigned | V1Item32Bits) /**< uint32       */
	V1ItemInt64  = V1PackItemType(V1ItemSigned | V1Item64Bits)   /**< int64       */
	V1ItemUint64 = V1PackItemType(V1ItemUnsigned | V1Item64Bits) /**< uint64       */
)

const (
	V2PackInvalid     = V2PackType(0x0)
	V2PackObject      = V2PackType(0x10)
	V2PackArray       = V2PackType(0x20)
	V2PackString      = V2PackType(0x50)
	V2PackRaw         = V2PackType(0x60)
	V2PackInt8        = V2PackType(0x11)
	V2PackInt16       = V2PackType(0x12)
	V2PackInt32       = V2PackType(0x14)
	V2PackInt64       = V2PackType(0x18)
	V2PackUint8       = V2PackType(0x21)
	V2PackUint16      = V2PackType(0x22)
	V2PackUint32      = V2PackType(0x24)
	V2PackUint64      = V2PackType(0x28)
	V2PackBool        = V2PackType(0x31)
	V2PackFloat       = V2PackType(0x44)
	V2PackDouble      = V2PackType(0x48)
	V2PackDate        = V2PackType(0x58)
	V2PackNull        = V2PackType(0x61)
	V2PackShortItem   = V2PackType(0x80)
	V2PackFixedItem   = V2PackType(0x0f)
	V2PackDeletedItem = V2PackType(0x70)
)

// V2Array represents array object
type V2Array []interface{}

// Size of the array
func (a V2Array) Size() int {
	return len(a)
}

type v2PackImpl struct {
	itemType byte
	name     []byte
	data     []byte
	items    []*v2PackImpl
	props    map[string]*v2PackImpl
}

func (p *v2PackImpl) isDeleted() bool {
	return !(p.itemType&byte(V2PackDeletedItem) > 0)
}
func (p *v2PackImpl) isShort() bool {
	return p.itemType&byte(V2PackShortItem) > 0
}
func (p *v2PackImpl) isFixed() bool {
	return p.itemType&byte(V2PackFixedItem) > 0
}
func (p *v2PackImpl) isArray() bool {
	return (p.itemType == byte(V2PackArray)) || (p.itemType == byte(V1PackArray))
}
func (p *v2PackImpl) isObject() bool {
	return (p.itemType == byte(V2PackObject)) || (p.itemType == byte(V1PackObject))
}
func (p *v2PackImpl) size() int {
	// TODO: check deleted(and possibly invalid, null) items?
	if p.isFixed() {
		return len(p.name) + len(p.data) + 1 + 1
	}
	if p.isShort() {
		return len(p.name) + len(p.data) + 1 + 1 + 1
	}
	return len(p.name) + len(p.data) + 1 + 1 + 4
}

// bind only result in ArrayObject or MapObject
func bind(pack *v2PackImpl) (interface{}, error) {
	var res interface{}

	if pack.isArray() {
		resArr := make(V2Array, 0)
		for _, item := range pack.items {
			sb, err := bind(item)
			if err != nil {
				return nil, err
			}
			resArr = append(resArr, sb)
		}
		res = resArr
		return res, nil
	}
	if pack.isObject() {
		resMap := make(V2Map)
		for name, item := range pack.props {
			sb, err := bind(item)
			if err != nil {
				return nil, err
			}
			resMap[name] = sb
		}
		res = resMap
		return res, nil
	}

	if V2PackType(pack.itemType)&V2PackString == V2PackString {
		if len(pack.data) > 0 && pack.data[len(pack.data)-1] == 0 {
			return string(pack.data[:len(pack.data)-1]), nil
		}
		return string(pack.data), nil
	} else if V2PackType(pack.itemType) == V2PackNull {
		return nil, nil
	} else if V2PackType(pack.itemType)&V2PackRaw == V2PackRaw {
		return pack.data[:], nil
	} else if V2PackType(pack.itemType)&V2PackInt8 == V2PackInt8 {
		return int8(pack.data[0]), nil
	} else if V2PackType(pack.itemType)&V2PackInt16 == V2PackInt16 {
		return int16(binary.LittleEndian.Uint16(pack.data)), nil
	} else if V2PackType(pack.itemType)&V2PackInt32 == V2PackInt32 {
		return int32(binary.LittleEndian.Uint32(pack.data)), nil
	} else if V2PackType(pack.itemType)&V2PackInt64 == V2PackInt64 {
		return int64(binary.LittleEndian.Uint64(pack.data)), nil
	} else if V2PackType(pack.itemType)&V2PackUint8 == V2PackUint8 {
		return uint8(pack.data[0]), nil
	} else if V2PackType(pack.itemType)&V2PackUint16 == V2PackUint16 {
		return binary.LittleEndian.Uint16(pack.data), nil
	} else if V2PackType(pack.itemType)&V2PackUint32 == V2PackUint32 {
		return binary.LittleEndian.Uint32(pack.data), nil
	} else if V2PackType(pack.itemType)&V2PackUint64 == V2PackUint64 {
		return binary.LittleEndian.Uint64(pack.data), nil
	} else if V2PackType(pack.itemType)&V2PackBool == V2PackBool {
		for _, b := range pack.data {
			if b > 0 {
				return true, nil
			}
		}
		return false, nil
	} else if V2PackType(pack.itemType)&V2PackFloat == V2PackFloat {
		return math.Float32frombits(binary.LittleEndian.Uint32(pack.data)), nil
	} else if V2PackType(pack.itemType)&V2PackDouble == V2PackDouble {
		return math.Float64frombits(binary.LittleEndian.Uint64(pack.data)), nil
	} else if V2PackType(pack.itemType)&V2PackNull == V2PackNull {
		return nil, nil
	} else {
		return nil, fmt.Errorf("invalid pack type 0x%x", pack.itemType)
	}
}

type v2PackIdx struct {
	packFullBytes []byte
	items         map[int]*v2PackImpl
	root          *v2PackImpl
}

func (idx *v2PackIdx) buildBufferFromObject(obj V2Map) (err error) {
	// idx.items = make(map[int]*v2PackImpl)

	idx.root, idx.packFullBytes, err = idx.writeObject(obj)

	return err
}

func (idx *v2PackIdx) writeObject(obj interface{}) (*v2PackImpl, []byte, error) {
	var pack v2PackImpl
	switch obj.(type) {
	case V2Map:
	case V2Array:
	case []byte:
	case string:
	case int64:
	case uint64:
	case int32:
	case uint32:
	case float32:
	case float64:
	case byte:
	case bool:
	}
	return &pack, nil, nil
}

func (idx *v2PackIdx) buildItemsFromBuffer(pack []byte) (err error) {
	idx.packFullBytes = pack
	idx.items = make(map[int]*v2PackImpl)
	idx.root, err = idx.readObjectAt(idx.packFullBytes, len(idx.packFullBytes), 0)
	return err
}
func (idx *v2PackIdx) readObjectAt(buf []byte, bufLength, offset int) (*v2PackImpl, error) {
	var pack v2PackImpl

	advanceOffset := func(step int) error {
		if offset+step > bufLength {
			return InvalidPackError
		}
		offset += step
		return nil
	}

	if err := advanceOffset(1); err != nil {
		return nil, err
	}
	pack.itemType = buf[offset-1]

	if err := advanceOffset(1); err != nil {
		return nil, err
	}
	packNameLength := uint8(buf[offset-1])

	if pack.isDeleted() {
		return nil, InvalidPackError
	}
	var contentSize uint32 = 0
	if pack.isFixed() {
		contentSize = uint32(pack.itemType & byte(V2PackFixedItem))
	} else if pack.isShort() {
		if err := advanceOffset(1); err != nil {
			return nil, err
		}
		contentSize = uint32(buf[offset-1])
	} else {
		if err := advanceOffset(4); err != nil {
			return nil, err
		}
		contentSize = binary.LittleEndian.Uint32(buf[offset-4 : offset])
	}

	if err := advanceOffset(int(packNameLength)); err != nil {
		return nil, err
	}
	pack.name = buf[offset-int(packNameLength) : offset]
	pack.data = buf[offset : offset+int(contentSize)]

	idx.items[offset] = &pack

	// should dive further into the data slice
	if pack.isArray() {
		count := binary.LittleEndian.Uint32(buf[offset : offset+4])
		pack.items = make([]*v2PackImpl, 0)
		for innerOff := offset + 4; innerOff < offset+int(contentSize); {
			obj, err := idx.readObjectAt(buf, bufLength, innerOff)
			if err != nil {
				return nil, err
			}
			pack.items = append(pack.items, obj)
			innerOff += obj.size()
		}
		if uint32(len(pack.items)) != count {
			return nil, InvalidPackError
		}
	} else if pack.isObject() {
		count := binary.LittleEndian.Uint32(buf[offset : offset+4])
		pack.props = make(map[string]*v2PackImpl)
		for innerOff := offset + 4; innerOff < offset+int(contentSize); {
			obj, err := idx.readObjectAt(buf, bufLength, innerOff)
			if err != nil {
				return nil, err
			}

			if len(obj.name) > 0 && obj.name[len(obj.name)-1] == 0 {
				pack.props[string(obj.name[:len(obj.name)-1])] = obj
			} else {
				pack.props[string(obj.name)] = obj
			}
			innerOff += obj.size()
		}
		if uint32(len(pack.props)) != count {
			return nil, InvalidPackError
		}
	}

	return &pack, nil
}
