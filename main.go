package main

import "flag"
import "fmt"
import "os"
import "os/exec"
import "time"

func main() {
	fmt.Println("Welcome to the redis challenge!")
	fmt.Println("")
	binaryPathPtr := flag.String("binary-path", "", "path to the redis executable to test. Ex: ./run_redis.sh")

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

	time.Sleep(1 * time.Second)

	if cmd.Process != nil {
		cmd.Process.Kill()
	}
	fmt.Println("Waiting for process to exit", *binaryPathPtr)
	cmd.Wait()
	fmt.Println("Tests done")
}

func runBinary(binaryPath string) (*exec.Cmd, error) {
	command := exec.Command(binaryPath)
	err := command.Start()
	if err != nil {
		return nil, err
	}

	return command, nil
}
