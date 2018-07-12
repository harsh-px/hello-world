package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kubernetes/kubernetes/pkg/kubectl/cmd"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func initK8sClient() (*kubernetes.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		return nil, fmt.Errorf("kubeconfig is not set")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Quick validation if client connection works
	_, err = client.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to k8s server: %s", err)
	}

	return client, nil
}

func PrintPodLogs(podName, podNamespace string) error {
	client, err := initK8sClient()
	if err != nil {
		return err
	}
}

func RunCommandInPod(cmds []string, podName, containerName, namespace string, stdout, stderr io.Writer, in io.Reader) (string, error) {
	client, err := initK8sClient()
	if err != nil {
		return "", err
	}

	pod, err := client.Core().Pods(namespace).Get(podName, meta_v1.GetOptions{})
	if err != nil {
		return "", err
	}

	if len(containerName) == 0 {
		if len(pod.Spec.Containers) != 1 {
			return "", fmt.Errorf("could not determine which container to use")
		}

		containerName = pod.Spec.Containers[0].Name
	}

	req := client.Core().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)

	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   cmds,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	execOpts := &cmd.ExecOptions{
		StreamOptions: cmd.StreamOptions{
			In:  in,
			Out: stdout,
			Err: stderr,
		},
		Executor: &cmd.DefaultRemoteExecutor{},
	}

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to init executor: %v", err)
	}

	fn := func() error {
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  in,
			Stdout: stdout,
			Stderr: stderr,
			Tty:    true,
		})
		if err != nil {
			return fmt.Errorf("could not execute: %v", err)
		}

		return nil
	}

	return "", fn()
}

//func demoRunPodCmds(podname, podnamespace string, cmds []string, in *bufio.Reader) (string, error) {
func demoRunPodCmds(podname, podnamespace string, cmds []string, stdout, stderr io.Writer, stdin io.Reader) (string, error) {
	fmt.Printf("[debug] foo...\n")
	/*input, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}

	fmt.Printf("input: %s\n", input)*/
	output, err := RunCommandInPod(cmds, podname, "", podnamespace, stdout, stderr, stdin)
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
