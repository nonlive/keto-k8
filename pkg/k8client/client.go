package k8client

import (
	"os/exec"
	"strings"
	"fmt"
	log "github.com/Sirupsen/logrus"

)

const cmdKubectl string = "kubectl"

// TODO: Use API - parse types from input YAML and create generic function
// Apply will take a yaml string and deploy it to the API...
func Apply(resource string) (error) {
	var args = []string {
		"apply",
		"-f",
	    "-",
	}

	output, err :=	runKubectl(args, resource)
	if err != nil {
		return fmt.Errorf("Error running kubectl:%s", output)
	}
	return nil
}

func runKubectl(cmdArgs []string, stdIn string) (out string, err error) {
	var cmdOut []byte

	cmdName := cmdKubectl
	log.Printf("Running:%v %v", cmdName, strings.Join(cmdArgs, " "))
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = strings.NewReader(stdIn)
	if cmdOut, err = cmd.CombinedOutput(); err != nil {
		return string(cmdOut[:]), err
	}
	return string(cmdOut[:]), nil
}
