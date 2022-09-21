package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

//getForcedDependencies returns the forcedDependency from the current file path
func getForcedDependencies(filePath string) (*ForcedDependencies, error) {
	var dependencies ForcedDependencies
	dependenciesFileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(dependenciesFileData, &dependencies); err != nil {
		return nil, err
	}
	return &dependencies, nil
}

//deleteVendor delete vendor Path.
func deleteVendor(directoryPath string) error {
	//delete vendor
	return os.RemoveAll(directoryPath + "/vendor")
}

//updateGoMod update current package go.mod file dependencies.
func updateGoMod(directoryPath string) error {
	//execute go mod vendor
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = directoryPath
	fmt.Printf("go mod tidy | path : %s \n", directoryPath)
	return cmd.Run()
}

//goModUpdateDependencies update/create vendor.
func updateVendor(directoryPath string) error {
	//execute go mod vendor
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Dir = directoryPath
	fmt.Printf("go mod vendor | path : %s\n", directoryPath)
	return cmd.Run()
}

func ConsolePopupRequest(question string, answers ...string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(question)
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		responseToLower := strings.ToLower(text)

		for _, answer := range answers {
			if responseToLower == strings.ToLower(answer) {
				return answer, nil
			}
		}
	}
}

//ExecuteCommandAndGetValue executes a Shel command and returns the output.
func ExecuteCommandAndGetValue(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
