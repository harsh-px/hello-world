package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/portworx/sched-ops/k8s"
	"github.com/sirupsen/logrus"
)

//func demoRunPodCmds(podname, podnamespace string, cmds []string, in *bufio.Reader) (string, error) {
func demoRunPodCmds(podname, podnamespace string, cmds []string, stdout, stderr io.Writer, stdin io.Reader) (string, error) {
	fmt.Printf("[debug] foo...\n")
	inst := k8s.Instance()
	/*input, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}

	fmt.Printf("input: %s\n", input)*/
	output, err := inst.RunCommandInPod(cmds, podname, "", podnamespace, stdout, stderr, stdin)
	if err != nil {
		return "", err
	}

	return output, nil
}

func main() {
	/*var kubeconfig string
	var node string
	var scname string*/
	var podname string

	//flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	/*flag.StringVar(&kubeconfig, "kubeconfig", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to kubeconfig.")
	flag.StringVar(&node, "node", "", "")
	flag.StringVar(&scname, "scname", "", "")
	flag.StringVar(&podname, "podname", "", "")
	flag.Parse()*/

	podname = "mysql-5499d4cd95-7c4s5"

	var output string
	var err error
	/*info, _ := os.Stdin.Stat()
	if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage:")
		fmt.Println("  cat yourfile.txt | searchr -pattern=<your_pattern>")
	} else if info.Size() > 0 {*/
	//fmt.Printf("using pipe mode...\n")
	//reader := bufio.NewReader(os.Stdin)
	output, err = demoRunPodCmds(podname, "default", []string{"bash"}, os.Stdout, os.Stderr, os.Stdin)
	if err != nil {
		logrus.Fatalf("failed with : %v", err)
	}
	//	}

	fmt.Printf("stdout: %s\n", output)
	time.Sleep(5 * time.Second)
}
