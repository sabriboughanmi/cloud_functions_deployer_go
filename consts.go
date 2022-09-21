package main

import "time"

var (
	deployPrefix = []string{"@AutoDeploy", "@autodeploy", "@Autodeploy", "@autoDeploy", "@AUTODEPLOY"}
	deploySuffix = []string{"@End", "@end", "@END"}
	deployArgs   = []string{"@Args:", "@args:", "Args:", "args:"}
)

const cloudFunctionsWriteLimit = time.Second * 10

const forcedDependenciesFileName = "forced_dependencies.json"

