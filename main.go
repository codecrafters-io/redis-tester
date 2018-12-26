package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "syscall"
import "time"
import "os/signal"

func main() {
	fmt.Println("Welcome to the redis challenge!")
	fmt.Println("")
	binaryPathPtr := flag.String(
		"binary-path",
		"",
		"path to the redis executable to test. Ex: ./run_redis.sh")

	debugPtr := flag.Bool(
		"debug",
		false,
		"Whether debug logs must be printed")

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
		fmt.Println("The --binary-path flag must be specified")
		os.Exit(1)
	}

	fmt.Println("Binary Path =", *binaryPathPtr)
	fmt.Println("Debug =", *debugPtr)
	fmt.Println("Report On Success =", *reportOnSuccessPtr)
	fmt.Println("Stage =", *currentStagePtr)
	fmt.Println("")

	cmd, err := runBinary(*binaryPathPtr, *debugPtr)
	if err != nil {
		fmt.Printf("Error when starting process: %s", err)
		fmt.Println("")
		os.Exit(1)
	}
	defer killCmdAndExit(cmd, 0)
	installSignalHandler(cmd)

	// TODO: Make this a proper wait?
	time.Sleep(1 * time.Second)

	result := newStageRunner(*debugPtr).Run(*currentStagePtr)
	if result.IsSuccess() {
		fmt.Println("")
		fmt.Println("All tests ran successfully. Congrats!")
		fmt.Println("")
	} else {
		killCmdAndExit(cmd, 1)
		return
	}

	if *reportOnSuccessPtr {
		report(result)
	} else {
		fmt.Println("If you'd like to report these results, " +
			"add the --report flag")
	}
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
		fmt.Printf("Error when killing process with PID %d: %s\n", cmd.Process.Pid, err)
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
