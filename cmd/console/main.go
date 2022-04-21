package main

import (
	"fmt"
	"github.com/mehdibo/godeploy/cmd/console/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
