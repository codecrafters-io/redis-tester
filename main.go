package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "syscall"
import "time"

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

	flag.Parse()

	if *binaryPathPtr == "" {
		fmt.Println("The --binary-path flag must be specified")
		os.Exit(1)
	}
	fmt.Println("Binary Path =", *binaryPathPtr)
	fmt.Println("Debug =", *debugPtr)
	fmt.Println("")

	cmd, err := runBinary(*binaryPathPtr, *debugPtr)
	if err != nil {
		fmt.Printf("Error when starting process: %s", err)
		fmt.Println("")
		os.Exit(1)
	}
	defer killCmdAndExit(cmd, 0)

	// TODO: Make this a proper wait?
	time.Sleep(1 * time.Second)

	result := newStageRunner().Run()
	if result.IsSuccess() {
		fmt.Println("All tests ran successfully.")
	} else {
		fmt.Println("Failed to run stage 1 test")
		fmt.Println(result.error)
		killCmdAndExit(cmd, 1)
	}

	fmt.Println("Tests done")
}

func killCmdAndExit(cmd *exec.Cmd, code int) {
	fmt.Printf("Killing process (PID: %d)\n", cmd.Process.Pid)
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		fmt.Printf("Error when killing process: %s\n", err)
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
