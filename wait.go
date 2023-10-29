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
		seoul := time.Now().In(loc)
		fmt.Println(seoul, "Waiting")
		time.Sleep(10 * time.Second)
	}
}