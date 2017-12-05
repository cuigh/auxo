// Package guid is a globally unique id generator.
// Code inspired from mgo/bson ObjectId.
package guid

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

const chars = "0123456789abcdefghijklmnopqrstuv"

var (
	// counter is atomically incremented when generating a new GUID
	// using New() function. It's used as a counter part of an id.
	counter uint32

	// machineID stores machine id generated once and used in subsequent calls
	// to NewObjectId function.
	machineID = readMachineID()
)

type ID [12]byte

// New returns a new unique id, see mgo package
func New() (id ID) {
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(id[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	id[4] = machineID[0]
	id[5] = machineID[1]
	id[6] = machineID[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	id[7] = byte(pid >> 8)
	id[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&counter, 1)
	id[9] = byte(i >> 16)
	id[10] = byte(i >> 8)
	id[11] = byte(i)
	return
}

// Parse parses a ID string.
//func Parse(s string) (ID, error) {
//	// todo:
//}

// String encodes ID as base32 string.
func (id ID) String() string {
	return string(id.encode())
}

// Slice encodes ID with base32.
func (id ID) Slice() []byte {
	return id[:]
}

// MarshalText implements encoding/text TextMarshaler interface
func (id ID) MarshalText() ([]byte, error) {
	return id.encode(), nil
}

func (id ID) encode() []byte {
	b := make([]byte, 20)

	b[0] = chars[id[0]>>3]
	b[1] = chars[(id[1]>>6)&0x1F|(id[0]<<2)&0x1F]
	b[2] = chars[(id[1]>>1)&0x1F]
	b[3] = chars[(id[2]>>4)&0x1F|(id[1]<<4)&0x1F]
	b[4] = chars[id[3]>>7|(id[2]<<1)&0x1F]
	b[5] = chars[(id[3]>>2)&0x1F]
	b[6] = chars[id[4]>>5|(id[3]<<3)&0x1F]
	b[7] = chars[id[4]&0x1F]
	b[8] = chars[id[5]>>3]
	b[9] = chars[(id[6]>>6)&0x1F|(id[5]<<2)&0x1F]
	b[10] = chars[(id[6]>>1)&0x1F]
	b[11] = chars[(id[7]>>4)&0x1F|(id[6]<<4)&0x1F]
	b[12] = chars[id[8]>>7|(id[7]<<1)&0x1F]
	b[13] = chars[(id[8]>>2)&0x1F]
	b[14] = chars[(id[9]>>5)|(id[8]<<3)&0x1F]
	b[15] = chars[id[9]&0x1F]
	b[16] = chars[id[10]>>3]
	b[17] = chars[(id[11]>>6)&0x1F|(id[10]<<2)&0x1F]
	b[18] = chars[(id[11]>>1)&0x1F]
	b[19] = chars[(id[11]<<4)&0x1F]

	return b
}

// readMachineID generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func readMachineID() []byte {
	id := make([]byte, 3)
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			panic(fmt.Errorf("cannot get hostname: %v; %v", err1, err2))
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	return id
}
