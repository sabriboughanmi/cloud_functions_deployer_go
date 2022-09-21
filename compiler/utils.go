package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/sabriboughanmi/go_utils/utils"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func isGoFile(file os.FileInfo) bool {
	if file.IsDir() {
		return false
	}
	return strings.HasSuffix(file.Name(), ".go")
}

//DeployFunctions deploys all tagged Cloud functions with associated CF parameters under the specified package path.
func DeployFunctions(projectPath string) error {
	path := projectPath
	fmt.Println("current Path: " + path)

	/*
		//execute go mod vendor
		var gcloudProjectID, err = ExecuteCommandAndGetValue("gcloud", "config", "get-value", "project")
		if err != nil {
			return err
		}

		answer, err := ConsolePopupRequest(fmt.Sprintf("Start Deploying Cloud Functions For Project : %s \n Start Deploy : Y,\n Cancel : N", gcloudProjectID), "y", "n")

		if answer == "n" {
			return fmt.Errorf("Deploy Rejected ! ")
		}
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

	*/

	packagesCommands, err := fetchDeployCommands(path)
	if err != nil {
		return err
	}

	processDuration := time.Now()
	var functionsCount = 0

	//Get global dependencies
	var globalPackageDependencies *ForcedDependencies
	globalPackageDependencies, err = getForcedDependencies(projectPath + "/" + forcedDependenciesFileName)

	//ignore error if global dependencies file doesn't exist
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	var errorChannel = make(chan error)
	var wg sync.WaitGroup
	//var rateLimiter = utils.CreateLimiter(cloudFunctionsWriteLimit, 5, len(packagesCommands))
	//rateLimiter.Start()

	for _, packageConfig := range packagesCommands {
		//Manage package dependencies

		//Pick the right package dependency
		var packageDependencies *ForcedDependencies

		//Check if package has its own dependencies
		if packageConfig.Dependencies != nil {
			packageDependencies = packageConfig.Dependencies
		} else {
			packageDependencies = globalPackageDependencies
		}

		//Update Vendor
		if err = updateGoMod(packageConfig.PackagePath); err != nil {
			return err
		}

		//Update dependencies if required.
		if packageDependencies != nil {
			if packageDependencies.RequireVendor {
				//Update Vendor
				if err = updateVendor(packageConfig.PackagePath); err != nil {
					return err
				}
			}

		}

		//Deploy Commands
		for _, functionCommand := range packageConfig.Commands {
			functionsCount++
			wg.Add(1)

			go func(fConfig command, packagePath string, waitGroup *sync.WaitGroup, errorChan chan error, functionIndex int) {
				defer waitGroup.Done()

				fConfig.Clean()

				cmd := exec.Command(fConfig[0], fConfig[1:]...)

				//Move to correct package directory before deploy
				cmd.Dir = packagePath

				// fmt.Printf("command: %v", args)
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				cmd.Stdout = nil

				//wait for Rate Limit
				//rateLimiter.Wait()
				time.Sleep(cloudFunctionsWriteLimit * time.Duration(functionIndex))

				fmt.Printf("Deploying %s Started! with command : %v\n", fConfig[3], fConfig)

				//Start command

				if err = cmd.Run(); err != nil {
					errorString := fmt.Sprintf("Deploying %s Failed with Error: %v, \n", fConfig[3], stderr.String())
					fmt.Println(errorString)
					errorChan <- fmt.Errorf(errorString)
					return
				}

				fmt.Printf("Deploying %s from path %s Completed!\n", fConfig[3], packagePath)
			}(functionCommand, packageConfig.PackagePath, &wg, errorChannel, functionsCount)

		}

	}

	//wait for channels
	wg.Wait()

	//Stop the Rate Limiter
	//rateLimiter.Stop()

	//Remove Vendors
	for _, packageCmds := range packagesCommands {
		if err = deleteVendor(packageCmds.PackagePath); err != nil {
			fmt.Printf("Error Deleting Vendor for Package : %s, with Error: %v !\n", packageCmds.PackagePath, err)
		}
	}

	var receivedErrors []string
	close(errorChannel)
	for err := range errorChannel {
		receivedErrors = append(receivedErrors, err.Error())
	}

	if len(receivedErrors) > 0 {
		return fmt.Errorf("Got %d Errors while commit - Errors : %s  \n", len(receivedErrors), string(utils.UnsafeAnythingToJSON(receivedErrors)))
	}

	fmt.Printf("  Deploying %d CloudFunctions Took: %v\n", functionsCount, time.Since(processDuration))
	fmt.Println("Deploy Completed Successfully!")
	fmt.Println("Click Enter to Exit")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	return nil
}

//fetchDeployCommands returns the GCloud deployment commands tagged under a package path, and it's subpackages.
func fetchDeployCommands(packagePath string) ([]packageCommands, error) {
	files, err := ioutil.ReadDir(packagePath)
	if err != nil {
		log.Fatal(err)
	}

	var packagesCommands []packageCommands

	var currentPackageCommands = packageCommands{
		Commands:     nil,
		PackagePath:  packagePath,
		Dependencies: nil,
	}

	//Check if current package has its own forced dependencies
	for _, file := range files {
		if file.Name() == forcedDependenciesFileName {
			var dependencies, err = getForcedDependencies(packagePath + "/" + file.Name())
			if err != nil {
				return nil, err
			}
			currentPackageCommands.Dependencies = dependencies
		}
	}

	var (
		deployCommandsPrefix = []string{"gcloud", "functions", "deploy"}
	)

	for _, file := range files {

		if file.IsDir() {
			//If vendor or Deprecated Skip
			if file.Name() == "vendor" || strings.Contains(strings.ToLower(file.Name()), "deprecated") {
				currentPackageCommands.RequireVendor = true
				continue
			}

			subCommands, err := fetchDeployCommands(fmt.Sprintf("%s/%s", packagePath, file.Name()))
			if err != nil {
				return nil, err
			}

			//Append deploy Commands
			if len(subCommands) > 0 {
				packagesCommands = append(packagesCommands, subCommands...)
			}
			continue
		}

		//Skip non Go Files
		if !isGoFile(file) {
			continue
		}

		goFile, err := os.OpenFile(packagePath+"/"+file.Name(), os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatalf("open file error: %v", err)
			return nil, err
		}
		//Collect Commands
		if err = func(f *os.File) error {

			//Defer for close
			defer f.Close()

			//lookingForNewCommand is false only when a command is unterminated. (command on Multiple Lines)
			var lookingState = lookingForNewCommand
			var currentDeployComment = ""

			sc := bufio.NewScanner(f)
			for sc.Scan() {
				textLine := sc.Text()

				switch lookingState {
				case lookingForNewCommand:
					//Check if textLine contains a new command
					if containsDeployCommand(textLine) != "" {
						currentDeployComment = textLine
						lookingState = lookingForCommentAnd
					}

					//Check if textLine contains and end command
					if containsEndDeployCommand(textLine) != "" {
						currentDeployComment += textLine
						lookingState = lookingForCloudFunction
					}
					continue
				case lookingForCommentAnd:
					//Check if textLine contains and end command
					currentDeployComment += textLine
					if containsEndDeployCommand(textLine) != "" {
						lookingState = lookingForCloudFunction
					}
					continue
				case lookingForCloudFunction:
					//Check for the nearest Cloud Function
					if funcName := getFunctionName(textLine); funcName != "" {

						args, err := parseComment(currentDeployComment)
						if err != nil {
							return err
						}
						var newCommand = append(deployCommandsPrefix, append([]string{funcName}, args...)...)
						currentPackageCommands.Commands = append(currentPackageCommands.Commands, newCommand)

						//reset Deploy comment as a new command can be found in the same file
						currentDeployComment = ""

						//turn back looking for new Commands
						lookingState = lookingForNewCommand
					}
					continue
				}

			}
			if err := sc.Err(); err != nil {
				log.Fatalf("scan file error: %v", err)
				return err
			}

			return nil
		}(goFile); err != nil {
			return nil, err
		}

	}
	//Add Package commands if deployment commands exists
	if currentPackageCommands.Commands != nil {
		packagesCommands = append(packagesCommands, currentPackageCommands)
	}

	return packagesCommands, nil
}
