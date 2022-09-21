package main

import (
	"errors"
	"fmt"
	"strings"
)

//parseCommandLine creates a cmd command from a string
func parseCommandLine(command string) (command, error) {
	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true
	for i := 0; i < len(command); i++ {
		c := command[i]

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return []string{}, errors.New(fmt.Sprintf("Unclosed quote in command line: %s", command))
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}

//parseComment returns the command from a Comment
func parseComment(comment string) (command, error) {

	depStartPrefix := containsDeployCommand(comment)
	depEndSuffix := containsEndDeployCommand(comment)
	depArgsPrefix := containsDeployArgs(comment)

	if depStartPrefix == "" || depEndSuffix == "" || depArgsPrefix == "" {
		return nil, fmt.Errorf("command is in Wrong format, "+
			"make sure The Deploy command is in the following format %s %s parametre1 parametre2 ... %s  : \n"+
			" Current command : %s\n", deployPrefix[0], deployArgs[0], deploySuffix[0], comment)
	}

	comment = strings.ReplaceAll(comment, "//", " ")
	comment = strings.ReplaceAll(comment, "\n", "")
	startIndex := strings.Index(comment, depArgsPrefix) + len(depArgsPrefix)
	endIndex := strings.Index(comment, depEndSuffix)
	return parseCommandLine(comment[startIndex:endIndex])
}

//containsDeployCommand returns the prefix if the comment  has a Deploy command
func containsDeployCommand(comment string) string {
	for _, prefix := range deployPrefix {
		if strings.Contains(comment, prefix) {
			return prefix
		}
	}
	return ""
}

//containsEndDeployCommand returns the suffix if the comment has an @End Deploy command
func containsEndDeployCommand(comment string) string {
	for _, suffix := range deploySuffix {
		if strings.Contains(comment, suffix) {
			return suffix
		}
	}
	return ""
}

//containsDeployArgs returns argPrefix if the comment contains Deployment Args
func containsDeployArgs(comment string) string {
	for _, argPrefix := range deployArgs {
		if strings.Contains(comment, argPrefix) {
			return argPrefix
		}
	}
	return ""
}

//getFunctionName returns the function name from the string.
func getFunctionName(textLine string) string {
	if strings.Contains(textLine, "func") {
		return strings.ReplaceAll(textLine[4:strings.Index(textLine, "(")], " ", "")
	}
	return ""
}

//cleanArg clears all useless spaces
func cleanArg(arg string) string {
	if strings.HasPrefix(arg, " ") {
		return cleanArg(arg[1:len(arg)])
	}
	if strings.HasSuffix(arg, " ") {
		return cleanArg(arg[0:len(arg)])
	}
	if strings.Contains(arg, "  ") {
		return cleanArg(strings.ReplaceAll(arg, "  ", " "))
	}
	return arg
}

//Clean removes all useless Spaces
func (command *command) Clean() {
	for i, arg := range *command {
		(*command)[i] = cleanArg(arg)
	}
}



