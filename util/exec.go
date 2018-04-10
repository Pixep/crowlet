package util

import (
	"bufio"
	"fmt"
	"os/exec"

	log "github.com/Sirupsen/logrus"
)

func Exec(cmdLine string, prefix string) {
	cmd := exec.Command(cmdLine)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("error creating stdout pipe for cmd", err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Info(prefix, "-cmd start")
			fmt.Printf("%s\n", scanner.Text())
			log.Info(prefix, "-cmd end")
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatalf("error starting cmd", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("error waiting for cmd", err)
	}
}
