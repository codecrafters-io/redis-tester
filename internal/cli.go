package internal

import (
	"fmt"
	"time"
)

func RunCLI(env map[string]string) int {
	context, err := GetContext(env)
	if err != nil {
		fmt.Printf("%s", err)
		return 1
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
		return 1
	}

	if antiCheatRunner().Run(executable).error != nil {
		return 1
	}

	fmt.Println("")
	fmt.Println("All tests ran successfully. Congrats!")
	fmt.Println("")

	// TODO: Print next stage!
	return 0
}
func runInOrder(runner StageRunner, executable *Executable) (StageRunnerResult, error) {
	result := runner.Run(executable)
	if !result.IsSuccess() {
		return result, fmt.Errorf("error")
	}

	return result, nil
}
