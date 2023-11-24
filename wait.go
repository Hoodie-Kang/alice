package main

import(
	"time"
	"fmt"
)

func main() {
	for {
		loc, err := time.LoadLocation("Asia/Seoul")
		if err != nil {
			panic(err)
		}
		now := time.Now()
		seoul := now.In(loc)
		fmt.Println(seoul, "Still Running")
		time.Sleep(10 * time.Second)
	}
}