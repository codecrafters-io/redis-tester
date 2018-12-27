package main

import "flag"
import "fmt"

// Context holds all flags that a user has passed in
type Context struct {
	binaryPath        string
	isDebug           bool
	currentStageIndex int
	reportOnSuccess   bool
	apiKey            string
}

func (c Context) print() {
	fmt.Println("Binary Path =", c.binaryPath)
	fmt.Println("Debug =", c.isDebug)
	fmt.Println("Report On Success =", c.reportOnSuccess)
	fmt.Println("Stage =", c.currentStageIndex)
}

// GetContext parses flags and returns a Context object
func GetContext() (Context, error) {
	binaryPathPtr := flag.String(
		"binary-path",
		"",
		"path to the redis executable to test. Ex: ./run_redis.sh")

	debugPtr := flag.Bool(
		"debug",
		false,
		"Whether debug logs must be printed")

	apiKeyPtr := flag.String(
		"api-key",
		"",
		"API key to use for reporting test results")

	reportOnSuccessPtr := flag.Bool(
		"report",
		false,
		"Whether test results must be reported")

	currentStagePtr := flag.Int(
		"stage",
		0,
		"The current stage you're on")

	flag.Parse()

	if *binaryPathPtr == "" {
		return Context{}, fmt.Errorf("" +
			"The --binary-path flag must be specified")
	}

	if *reportOnSuccessPtr && (*apiKeyPtr == "") {
		return Context{}, fmt.Errorf("" +
			"If --report is specified, " +
			"--api-key must be specified too.")
	}

	return Context{
		binaryPath:        *binaryPathPtr,
		isDebug:           *debugPtr,
		currentStageIndex: *currentStagePtr,
		reportOnSuccess:   *reportOnSuccessPtr,
		apiKey:            *apiKeyPtr,
	}, nil
}
