package util

import (
	"bufio"
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

// Exec runs a system command
func Exec(cmdLine string, prefix string) {
	log.Info("Running '", cmdLine, "'...")
	cmd := exec.Command(cmdLine)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Error("Error reading from command standard output:", err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	err = cmd.Start()
	if err != nil {
		log.Error("Failed to start command:", err)
		return
	}

	for scanner.Scan() {
		fmt.Printf("%s\n", scanner.Text())
	}

	err = cmd.Wait()
	if err != nil {
		log.Error("Error waiting for command:", err)
	}
}
