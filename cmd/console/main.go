package main

import (
	"fmt"
	"github.com/mehdibo/go_deploy/cmd/console/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
