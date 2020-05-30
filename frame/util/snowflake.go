package util

// Package snowflake implements Snowflake, a distributed unique ID generator inspired by Twitter's Snowflake.
//
// A Snowflake ID is composed of
//     39 bits for time in units of 10 msec
//      8 bits for a sequence number
//     16 bits for a machine id
import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// These constants are the bit lengths of Snowflake ID parts.
const (
	BitLenSequence  = 8  // bit length of sequence number
	BitLenBiz       = 4  // bit length of biz number
	BitLenMachineID = 16 // bit length of machine id
	BitLenTime      = 35 // bit length of time
)

// Settings configures Snowflake:
//
// MachineID returns the unique ID of the Snowflake instance.
// If MachineID returns an error, Snowflake is not created.
// If MachineID is nil, default MachineID is used.
// Default MachineID returns the lower 16 bits of the private IP address.
//
// CheckMachineID validates the uniqueness of the machine ID.
// If CheckMachineID returns false, Snowflake is not created.
// If CheckMachineID is nil, no validation is done.
type Settings struct {
	Biz            uint16                 //业务码
	MachineID      func() (uint16, error) //默认不使用
	CheckMachineID func(uint16) bool      //默认不使用
}

// Snowflake is a distributed unique ID generator.
type Snowflake struct {
	mutex       *sync.Mutex
	startTime   int64 //毫秒
	elapsedTime int64
	sequence    uint16
	machineID   uint16 //机器码 本机地址 后两位
	biz         uint16 //业务码
}

/*

|nouse (1) |time (35) |machineID (16)| biz (4)|seq (8)|
*/

// NewSnowflake returns a new Snowflake configured with the given Settings.
// NewSnowflake returns nil in the following cases:
// - Settings.StartTime is ahead of the current time.
// - Settings.MachineID returns an error.
// - Settings.CheckMachineID returns false.
func NewSnowflake(st Settings) *Snowflake {
	sf := new(Snowflake)
	sf.mutex = new(sync.Mutex)
	sf.sequence = 0
	sf.biz = st.Biz
	// 这个时间具有特殊意义
	sf.startTime = toSnowflakeTime(time.Date(2016, 5, 10, 0, 0, 0, 0, time.UTC))
	// sf.startTime = toSnowflakeTime(time.Now())
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
// After the Snowflake time overflows, NextID returns an error.
func (sf *Snowflake) NextID() (uint64, error) {
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

	return sf.toID()
}

const snowflakeTimeUnit = 1e7 // nsec, i.e. 10 msec

//to 毫秒
func toSnowflakeTime(t time.Time) int64 {
	return t.UTC().UnixNano() / snowflakeTimeUnit
}

//开始到现在已经过了多少毫秒
func currentElapsedTime(startTime int64) int64 {
	return toSnowflakeTime(time.Now()) - startTime
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%snowflakeTimeUnit)*time.Nanosecond
}

// |nouse (1) |time (35) |machineID (16)| biz (4)|seq (8)|
func (sf *Snowflake) toID() (uint64, error) {
	if sf.elapsedTime >= 1<<BitLenTime {
		return 0, errors.New("over the time limit")
	}
	return uint64(sf.elapsedTime)<<(BitLenSequence+BitLenBiz+BitLenMachineID) |
		uint64(sf.machineID)<<(BitLenSequence+BitLenBiz) |
		uint64(sf.biz)<<(BitLenSequence) |
		uint64(sf.sequence), nil
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

// |nouse (1) |time (35) |machineID (16)| biz (4)|seq (8)|
// Decompose returns a set of Snowflake ID parts.
func Decompose(id uint64) map[string]uint64 {
	const maskSequence = uint64((1<<BitLenSequence - 1))
	const maskBiz = uint64(1<<(BitLenBiz) - 1)
	const maskMachineID = uint64(1<<BitLenMachineID - 1)

	msb := id >> 63

	sequence := id & maskSequence
	biz := (id >> BitLenSequence) & maskBiz
	machineID := (id >> (BitLenSequence + BitLenBiz)) & maskMachineID
	time := id >> (BitLenSequence + BitLenMachineID + BitLenBiz)
	return map[string]uint64{
		"id":         id,
		"msb":        msb,
		"time":       time,
		"sequence":   sequence,
		"machine_id": machineID,
		"biz":        biz,
	}
}

//使用demo
func demo() {
	var st Settings
	st.Biz = 1
	sf := NewSnowflake(st)
	if sf == nil {
		panic("snowflake not created")
	}
	id, err := sf.NextID()
	fmt.Println(id, err)
	parts := Decompose(id)
	fmt.Println(parts)
}
