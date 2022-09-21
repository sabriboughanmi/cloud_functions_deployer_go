package main

type command []string

//packageCommands Contains a package commands, path and dependencies
type packageCommands struct {
	Commands     []command
	PackagePath  string
	Dependencies *ForcedDependencies
	RequireVendor bool
}

//ForcedDependencies specify a package dependencies.
type ForcedDependencies struct {
	RequireVendor bool
}

type ParsingState int

const (
	lookingForNewCommand    = 0 //Process is looking for a new Cloud Function
	lookingForCommentAnd    = 1 //Process is looking for the end of the command
	lookingForCloudFunction = 2 //Process is looking for the Cloud Function related to the command

)
