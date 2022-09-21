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
	//	packagePath = "C:\\Go Projects\\cwbr_v2\\cloudfunctions"

	fmt.Println(packagePath)

	if err := DeployFunctions(packagePath); err != nil {
		fmt.Println(err)
		return
	}

}
