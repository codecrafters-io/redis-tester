package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	context, err := GetContext(envMap())
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	if context.isDebug {
		context.print()
		fmt.Println("")
	}

	executable := NewVerboseExecutable(context.binaryPath, getLogger(true, "[your_program] ").Plainln)

	// TODO: Make this a proper wait?
	time.Sleep(1 * time.Second)

	runner := newStageRunner(context.isDebug)
	runner = runner.Truncated(context.currentStageSlug)

	_, err = runInOrder(runner, executable)
	if err != nil {
		os.Exit(1)
		return
	}

	if antiCheatRunner().Run(executable).error != nil {
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("All tests ran successfully. Congrats!")
	fmt.Println("")

	// TODO: Print next stage!
}

func envMap() map[string]string {
	result := make(map[string]string)
	for _, keyVal := range os.Environ() {
		split := strings.SplitN(keyVal, "=", 2)
		key, val := split[0], split[1]
		result[key] = val
	}

	return result
}

func runRandomizedMultipleAndLog(runner StageRunner, executable *Executable) error {
	fmt.Println("")
	fmt.Println("Running tests multiple times to make sure...")

	fmt.Println("")
	time.Sleep(1 * time.Second)

	for i := 1; i <= 2; i++ {
		fmt.Printf("Randomized Run #%d...\n\n", i)
		time.Sleep(1 * time.Second)
		err := runRandomized(runner, executable)
		if err != nil {
			return err
		}
		fmt.Println("")
	}

	return nil
}

func runInOrder(runner StageRunner, executable *Executable) (StageRunnerResult, error) {
	result := runner.Run(executable)
	if !result.IsSuccess() {
		return result, fmt.Errorf("error")
	}

	return result, nil
}

func runRandomized(runner StageRunner, executable *Executable) error {
	result := runner.Randomized().Run(executable)
	if !result.IsSuccess() {
		return fmt.Errorf("error")
	}

	return nil
}

// func installSignalHandler(cmd *exec.Cmd) {
// 	c := make(chan os.Signal, 1)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
// 	go func() {
// 		for range c {
// 			// sig is a ^C, handle it
// 			os.Exit(0)
// 		}
// 	}()
// }
