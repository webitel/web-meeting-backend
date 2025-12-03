package main

import (
	"fmt"

	"github.com/webitel/web-meeting-backend/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}
}
