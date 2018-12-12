package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "time"
import "github.com/go-redis/redis"
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

	// Run tests here
	time.Sleep(1 * time.Second)

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
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := command.Start()
	if err != nil {
		return nil, err
	}

	return command, nil
}

func runStage1() error {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6739",
	})
	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}

	if pong != "pong" {
		return fmt.Errorf("Expected pong, got %s", pong)
	}

	return nil
}
