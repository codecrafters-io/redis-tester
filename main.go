package main

import "fmt"
import "os"
import "os/exec"
import "syscall"
import "time"
import "os/signal"

func main() {
	fmt.Println("Welcome to the redis challenge!")
	fmt.Println("")

	context, err := GetContext()
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	context.print()
	fmt.Println("")

	cmd, err := runBinary(context.binaryPath, context.isDebug)
	if err != nil {
		fmt.Printf("Error when starting process: %s", err)
		fmt.Println("")
		os.Exit(1)
	}
	defer killCmdAndExit(cmd, 0)
	installSignalHandler(cmd)

	// TODO: Make this a proper wait?
	time.Sleep(1 * time.Second)

	runner := newStageRunner(context.isDebug)
	runner = runner.Truncated(context.currentStageIndex)

	result, err := runInOrder(runner)
	if err != nil {
		killCmdAndExit(cmd, 1)
		return
	}

	if !context.reportOnSuccess {
		fmt.Println("If you'd like to report these " +
			"results, add the --report flag")
		return
	}

	if context.currentStageIndex > 0 {
		err = runRandomizedMultipleAndLog(runner)
		if err != nil {
			killCmdAndExit(cmd, 1)
		}
	}

	if antiCheatRunner().Run().error != nil {
		killCmdAndExit(cmd, 1)
	}

	time.Sleep(1 * time.Second)
	if report(result, context.apiKey) != nil {
		killCmdAndExit(cmd, 1)
	}
}

func runRandomizedMultipleAndLog(runner StageRunner) error {
	fmt.Println("Running tests multiple times to make sure...")

	fmt.Println("")
	time.Sleep(1 * time.Second)

	for i := 1; i <= 5; i++ {
		fmt.Printf("%d...\n\n", i)
		time.Sleep(1 * time.Second)
		err := runRandomized(runner)
		if err != nil {
			return err
		}
		fmt.Println("")
	}

	return nil
}

func runInOrder(runner StageRunner) (StageRunnerResult, error) {
	result := runner.Run()
	if !result.IsSuccess() {
		return result, fmt.Errorf("error")
	}

	fmt.Println("")
	fmt.Println("All tests ran successfully. Congrats!")
	fmt.Println("")
	return result, nil
}

func runRandomized(runner StageRunner) error {
	result := runner.Randomized().Run()
	if !result.IsSuccess() {
		return fmt.Errorf("error")
	}

	return nil
}

func installSignalHandler(cmd *exec.Cmd) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			// sig is a ^C, handle it
			killCmdAndExit(cmd, 0)
		}
	}()
}

func killCmdAndExit(cmd *exec.Cmd, code int) {
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		fmt.Printf("Error when killing process "+
			"with PID %d: %s\n", cmd.Process.Pid, err)
	}
	os.Exit(code)
}

func runBinary(binaryPath string, debug bool) (*exec.Cmd, error) {
	command := exec.Command(binaryPath)
	if debug {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := command.Start()
	if err != nil {
		return nil, err
	}

	return command, nil
}
