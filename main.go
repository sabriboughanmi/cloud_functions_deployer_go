package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	//go list -u -m -json all | go-mod-outdated  -direct
	//Command above can be used to see dependency updates available per module.

	packagePath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	packagePath = "C:\\Users\\safab\\NewT4u\\cloudfunctions"

	fmt.Println(packagePath)

	if err := DeployFunctions(packagePath); err != nil {
		fmt.Println(err)
		return
	}
}
