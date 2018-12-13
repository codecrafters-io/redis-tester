package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "syscall"

func main() {
	fmt.Println("Welcome to the redis challenge!")
	fmt.Println("")
	binaryPathPtr := flag.String(
		"binary-path",
		"",
		"path to the redis executable to test. Ex: ./run_redis.sh")

	flag.Parse()
	if *binaryPathPtr == "" {
		fmt.Println("The --binary-path flag must be specified")
		os.Exit(1)
	}
	fmt.Println("Binary Path =", *binaryPathPtr)
	fmt.Println("")

	cmd, err := runBinary(*binaryPathPtr)
	if err != nil {
		fmt.Printf("Error when starting process: %s", err)
		fmt.Println("")
		os.Exit(1)
	}
	defer killCmdAndExit(cmd, 0)

	stageRunner := newStageRunner()
	stageRunner.Run()

	fmt.Println("Running stage 1 test...")
	if err := runStage1(); err != nil {
		fmt.Println("Failed to run stage 1 test")
		fmt.Println(err)
		killCmdAndExit(cmd, 1)
	}

	fmt.Println("Tests done")
}

func killCmdAndExit(cmd *exec.Cmd, code int) {
	fmt.Printf("Killing process %d", cmd.Process.Pid)
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		fmt.Printf("Error when killing process: %s\n", err)
	}
	os.Exit(code)
}

func runBinary(binaryPath string) (*exec.Cmd, error) {
	command := exec.Command(binaryPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := command.Start()
	if err != nil {
		return nil, err
	}

	return command, nil
}
