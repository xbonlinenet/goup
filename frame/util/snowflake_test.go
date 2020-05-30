package util

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set"
)

var sf *Snowflake

var machineID uint64

func init() {
	var st Settings
	st.Biz = 1
	sf = NewSnowflake(st)
	if sf == nil {
		panic("snowflake not created")
	}
	ip, _ := lower16BitPrivateIP()
	machineID = uint64(ip)
}

func nextID(t *testing.T) uint64 {
	id, err := sf.NextID()
	if err != nil {
		t.Fatal("id not generated err:", err)
	}
	return id
}

func TestSnowflakeOnce(t *testing.T) {
	sleepTime := uint64(50)
	time.Sleep(time.Duration(sleepTime) * 10 * time.Millisecond)

	id := nextID(t)
	parts := Decompose(id)

	fmt.Println("snowflake id:", id)
	fmt.Println("decompose:", parts)

	actualMSB := parts["msb"]
	if actualMSB != 0 {
		t.Errorf("unexpected msb: %d", actualMSB)
	}

	// start time = now
	// actualTime := parts["time"]
	// if (actualTime < sleepTime || actualTime > sleepTime+1) {
	// 	t.Errorf("unexpected time: %d", actualTime)
	// }

	actualSequence := parts["sequence"]
	if actualSequence != 0 {
		t.Errorf("unexpected sequence: %d", actualSequence)
	}

	actualMachineID := parts["machine_id"]
	if actualMachineID != machineID {
		t.Errorf("unexpected machine id: %d", actualMachineID)
	}

}

func currentTime() int64 {
	return toSnowflakeTime(time.Now())
}

func TestSnowflakeFor10Sec(t *testing.T) {
	var numID uint32
	var lastID uint64
	var maxSequence uint64

	initial := currentTime()
	current := initial
	for current-initial < 1000 {
		id := nextID(t)
		parts := Decompose(id)
		numID++

		if id <= lastID {
			t.Fatal("duplicated id")
		}
		lastID = id

		current = currentTime()

		actualMSB := parts["msb"]
		if actualMSB != 0 {
			t.Errorf("unexpected msb: %d", actualMSB)
		}

		actualSequence := parts["sequence"]
		if maxSequence < actualSequence {
			maxSequence = actualSequence
		}

		actualMachineID := parts["machine_id"]
		if actualMachineID != machineID {
			t.Errorf("unexpected machine id: %d", actualMachineID)
		}
	}

	if maxSequence != 1<<BitLenSequence-1 {
		t.Errorf("unexpected max sequence: %d", maxSequence)
	}
	fmt.Println("max sequence:", maxSequence)
	fmt.Println("number of id:", numID)
}

func TestSnowflakeInParallel(t *testing.T) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Println("number of cpu:", numCPU)

	consumer := make(chan uint64)

	const numID = 10000
	generate := func() {
		for i := 0; i < numID; i++ {
			consumer <- nextID(t)
		}
	}

	const numGenerator = 10
	for i := 0; i < numGenerator; i++ {
		go generate()
	}

	set := mapset.NewSet()
	for i := 0; i < numID*numGenerator; i++ {
		id := <-consumer
		if set.Contains(id) {
			t.Fatal("duplicated id")
		} else {
			set.Add(id)
		}
	}
	fmt.Println("number of id:", set.Cardinality())
}

func TestNilSnowflake(t *testing.T) {

	var noMachineID Settings
	noMachineID.MachineID = func() (uint16, error) {
		return 0, fmt.Errorf("no machine id")
	}
	if NewSnowflake(noMachineID) != nil {
		t.Errorf("snowflake with no machine id")
	}

	var invalidMachineID Settings
	invalidMachineID.CheckMachineID = func(uint16) bool {
		return false
	}
	if NewSnowflake(invalidMachineID) != nil {
		t.Errorf("snowflake with invalid machine id")
	}
}

func pseudoSleep(period time.Duration) {
	sf.startTime -= int64(period) / snowflakeTimeUnit
}

func TestNextIDError(t *testing.T) {
	// year := time.Duration(365*24) * time.Hour
	// pseudoSleep(time.Duration(174) * year)
	// nextID(t)

	// pseudoSleep(time.Duration(1) * year)
	// _, err := sf.NextID()
	// if err == nil {
	// 	t.Errorf("time is not over")
	// }
}
