package utils

import (
	"errors"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

/*
sonyflake 实现方式：
0 | 0000000  00000000  00000000  00000000  000000|   000000     | 00 00     | 00000000 00000000
  |           时间戳,精确到10毫秒,37位           |sequence(6位) |随机数(4位)| MachineID机器号（16位,默认用私有ip后两位|
*/
var flake = NewSonyflake(Settings{})

func GetUuidString() string {
	return strconv.FormatUint(GetUuidUInt64(), 10)
}

func GetUuidUInt64() uint64 {
	uid, _ := flake.NextID()
	return uid
}

// Package sonyflake implements Sonyflake, a distributed unique ID generator inspired by Twitter's Snowflake.
//
// A Sonyflake ID is composed of
//     37 bits for time in units of 10 msec
//      6 bits for a sequence number
//      4 bits for a random number
//     16 bits for a machine id

// These constants are the bit lengths of Sonyflake ID parts.
const (
	BitLenTime      = 37                                                 // bit length of time
	BitLenSequence  = 6                                                  // bit length of sequence number
	BitLenRandomNum = 4                                                  // bit length of random number
	BitLenMachineID = 63 - BitLenTime - BitLenRandomNum - BitLenSequence // bit length of machine id
)

// Settings configures Sonyflake:
//
// StartTime is the time since which the Sonyflake time is defined as the elapsed time.
// If StartTime is 0, the start time of the Sonyflake is set to "2014-09-01 00:00:00 +0000 UTC".
// If StartTime is ahead of the current time, Sonyflake is not created.
//
// MachineID returns the unique ID of the Sonyflake instance.
// If MachineID returns an error, Sonyflake is not created.
// If MachineID is nil, default MachineID is used.
// Default MachineID returns the lower 16 bits of the private IP address.
//
// CheckMachineID validates the uniqueness of the machine ID.
// If CheckMachineID returns false, Sonyflake is not created.
// If CheckMachineID is nil, no validation is done.
type Settings struct {
	StartTime      time.Time
	MachineID      func() (uint16, error)
	CheckMachineID func(uint16) bool
}

// Sonyflake is a distributed unique ID generator.
type Sonyflake struct {
	mutex       *sync.Mutex
	generator   *rand.Rand
	startTime   int64
	elapsedTime int64
	randomNum   uint16
	sequence    uint16
	machineID   uint16
}

// NewSonyflake returns a new Sonyflake configured with the given Settings.
// NewSonyflake returns nil in the following cases:
// - Settings.StartTime is ahead of the current time.
// - Settings.MachineID returns an error.
// - Settings.CheckMachineID returns false.
func NewSonyflake(st Settings) *Sonyflake {
	sf := new(Sonyflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = uint16(1<<BitLenSequence - 1)

	if st.StartTime.After(time.Now()) {
		return nil
	}
	if st.StartTime.IsZero() {
		// todo this start time should be the latest, but since there are already businesses using it, it can't be changed
		sf.startTime = toSonyflakeTime(time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC))
	} else {
		sf.startTime = toSonyflakeTime(st.StartTime)
	}

	sf.generator = NewRand(time.Now().UnixNano())

	var err error
	if st.MachineID == nil {
		sf.machineID, err = lower16BitPrivateIP()
	} else {
		sf.machineID, err = st.MachineID()
	}
	if err != nil || (st.CheckMachineID != nil && !st.CheckMachineID(sf.machineID)) {
		return nil
	}

	return sf
}

// NextID generates a next unique ID.
// After the Sonyflake time overflows, NextID returns an error.
func (sf *Sonyflake) NextID() (uint64, error) {
	const maskSequence = uint16(1<<BitLenSequence - 1)

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	current := currentElapsedTime(sf.startTime)
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else { // sf.elapsedTime >= current
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime((overtime)))
		}
	}

	sf.randomNum = uint16(sf.generator.Intn(2 << (BitLenRandomNum - 1)))
	return sf.toID()
}

const sonyflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

func toSonyflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / sonyflakeTimeUnit
}

func currentElapsedTime(startTime int64) int64 {
	return toSonyflakeTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%sonyflakeTimeUnit)*time.Nanosecond
}

func (sf *Sonyflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}

	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenRandomNum+BitLenMachineID) |
		uint64(sf.sequence)<<(BitLenRandomNum+BitLenMachineID) |
		uint64(sf.randomNum)<<(BitLenMachineID) |
		uint64(sf.machineID), nil
}

func privateIPv4() (net.IP, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}
	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}

func lower16BitPrivateIP() (uint16, error) {
	ip, err := privateIPv4()
	if err != nil {
		return 0, err
	}

	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}

// Decompose returns a set of Sonyflake ID parts.
func Decompose(id uint64) map[string]uint64 {
	const maskSequence = uint64((1<<BitLenSequence - 1) << (BitLenMachineID + BitLenRandomNum))
	const maskRandomNum = uint64((1<<BitLenRandomNum - 1) << (BitLenMachineID))
	const maskMachineID = uint64(1<<BitLenMachineID - 1)

	msb := id >> 63
	time := id >> (BitLenSequence + BitLenRandomNum + BitLenMachineID)
	sequence := id & maskSequence >> (BitLenRandomNum + BitLenMachineID)
	randomNum := id & maskRandomNum >> BitLenMachineID
	machineID := id & maskMachineID

	return map[string]uint64{
		"id":         id,
		"msb":        msb,
		"time":       time,
		"random-num": randomNum,
		"sequence":   sequence,
		"machine-id": machineID,
	}
}
