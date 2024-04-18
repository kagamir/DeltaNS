package common

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"
)

func CheckTime(timestampBytes []byte) bool {
	now := time.Now()
	nowTimestamp := now.Unix()

	reader := bytes.NewReader(timestampBytes)
	timestamp, err := binary.ReadVarint(reader)
	if err != nil {
		log.Println("timestampBytes Error: ", err)
		return false
	}
	log.Println("CheckTime", nowTimestamp, timestamp)
	if nowTimestamp-timestamp < 60 {
		return true
	} else {
		return false
	}

}

func GetTimeBytes() []byte { // 10
	now := time.Now()
	nowTimestamp := now.Unix()
	buffer := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buffer, nowTimestamp)
	return buffer
}
