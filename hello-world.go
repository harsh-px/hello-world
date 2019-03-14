package main

import (
	"fmt"
	"time"
)

// pod is a simple type to encapsulate a pod's uid and namespace
type Pod struct {
	Uid       string `json:"uid"`
	Namespace string `json:"namespace"`
}

// commandTask tracks pods where commands for a taskID might still be running
type CommandTask struct {
	TaskID string `json:"task_id"`
	Pods   []Pod  `json:"pods"`
}

/*func initK8sClient() (*kubernetes.Clientset, error) {
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
}*/

func main() {
	/*var kubeconfig string
	var node string
	var scname string*/
	//var podname string

	//flag.StringVar(&kubeconfig, "kubeconfig", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to kubeconfig.")
	//flag.StringVar(&node, "node", "", "")
	//flag.StringVar(&scname, "scname", "", "")

	// Reference: https://vmware.github.io/vsphere-storage-for-kubernetes/documentation/existing.html
	// Essentially doing below here
	// cat /sys/class/dmi/id/product_serial | sed -e 's/^VMware-//' -e 's/-/ /' | awk '{ print toupper($1$2$3$4 "-" $5$6 "-" $7$8 "-" $9$10 "-" $11$12$13$14$15$16) }'
	t := time.Now()
	fmt.Printf("current time is : %v\n", t)
	fmt.Printf("week day is : %v\n", t.Weekday().String())
	t, err := time.Parse(time.RFC1123, t.Format(time.RFC1123))
	if err != nil {
		fmt.Printf("[error] failed to parse: %v", err)
		return
	}
}
