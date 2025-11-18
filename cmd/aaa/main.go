package main

import (
	"fmt"
	"time"
)

func main() {

	fmt.Println("Hello", CollapseTimestampId())
	time.Sleep(time.Minute * 1)
	fmt.Println("Hello", CollapseTimestampId())
	time.Sleep(time.Minute * 1)
	fmt.Println("Hello", CollapseTimestampId())
	time.Sleep(time.Minute * 1)

}

func CollapseTimestampId() int64 {
	interval := (time.Minute * 1).Seconds() // 1 minutes

	now := time.Now().Unix()
	rounded := now - (now % int64(interval))

	return rounded
}
