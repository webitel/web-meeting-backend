package main

import (
	"fmt"

	"github.com/webitel/web-meeting-backend/cmd"
	_ "github.com/webitel/web-meeting-backend/internal/utils"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}
}
