package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"time"

	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/portworx/sched-ops/k8s"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cloudSnapshotInitialDelay = 5 * time.Second
	cloudSnapshotFactor       = 1.5
	cloudSnapshotSteps        = 360
)

type Source struct {
	Parent string
}

func testPointer(source *Source) {
	if source == nil {
		fmt.Printf("initializing source...\n")
		source = &Source{}
	}
}

func testBackoff() error {
	var cloudsnapBackoff = wait.Backoff{
		Duration: cloudSnapshotInitialDelay,
		Factor:   cloudSnapshotFactor,
		Steps:    cloudSnapshotSteps,
	}

	err := wait.ExponentialBackoff(cloudsnapBackoff, func() (bool, error) {
		fmt.Printf("retrying...")
		return false, nil
	})

	if err != nil {
		logrus.Errorf("backoff function failed: %v", err)
		return err
	}

	fmt.Printf("backoff function started")

	return nil
}

func gen(nums ...int) error {
	go func() {
		fmt.Printf("sleeping forever")
		time.Sleep(24 * time.Hour)
		fmt.Printf("out of sleep")
		return
	}()

	return nil
}

func testGoRoutines() error {
	err := gen(1, 2)
	if err != nil {
		return err
	}

	fmt.Printf("returning from test function")
	runtime.Goexit()
	return nil
}

func loadClientFromKubeconfig(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func demoDS() error {
	inst := k8s.Instance()
	return inst.ValidateDaemonSet("portworx", "kube-system", 5*time.Minute)
}

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

func demoPXApps(nodename, scname, kubeconfig string) error {
	inst := k8s.Instance()
	pods, err := inst.GetPodsUsingVolumePluginByNodeName(nodename, "kubernetes.io/portworx-volume")
	if err != nil {
		return err
	}

	fmt.Printf("found node: %s pods: %v\n", nodename, pods)

	pods, err = inst.GetPodsUsingVolumePlugin("kubernetes.io/portworx-volume")
	if err != nil {
		return err
	}

	fmt.Printf("found all pods: %v\n", pods)

	deps, err := inst.GetDeploymentsUsingStorageClass(scname)
	if err != nil {
		return err
	}

	fmt.Printf("found deps: %v", deps)

	client, err := loadClientFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}

	for _, d := range deps {
		dCopy, err := client.Apps().Deployments(d.Namespace).Get(d.Name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		fmt.Printf("fetched: %v", dCopy)
	}

	ss, err := inst.GetStatefulSetsUsingStorageClass(scname)
	if err != nil {
		return err
	}

	fmt.Printf("found ss: %v", ss)
	return nil
}

func demoDrain(node string) error {
	inst := k8s.Instance()
	pods, err := inst.GetPodsUsingVolumePluginByNodeName(node, "kubernetes.io/portworx-volume")
	if err != nil {
		fmt.Printf("error creating k8s instance: %v", err)
		return err
	}

	fmt.Println("found pods: ")
	for _, p := range pods {
		fmt.Printf("\t%s\n", p.Name)
	}

	// filer
	var podsToDrain []v1.Pod
	for _, p := range pods {
		if inst.IsPodBeingManaged(p) {
			podsToDrain = append(podsToDrain, p)
		} else {
			fmt.Printf("skipping pod: %s as it's not being managed\n", p.Name)
		}
	}

	err = inst.DrainPodsFromNode(node, podsToDrain, 3*time.Minute)
	if err != nil {
		fmt.Printf("error: failed to drain pods from node. err: %v", err)
		return err
	}

	/*for _, p := range pods {
		fmt.Printf("debug: wait for deletion of pod: %s\n", p.Name)
		err = inst.WaitForPodDeletion(p.Name, p.Namespace, 2*time.Minute)
		if err != nil {
			fmt.Printf("error: failed to drain pod: %s from node. err: %v\n", p.Name, err)
		}
	}*/

	err = inst.UnCordonNode(node)
	if err != nil {
		fmt.Printf("error: failed to uncordon node: %s err: %v", node, err)
		return err
	}

	return nil
}

func demoPVC(scname string) error {
	inst := k8s.Instance()

	pvcs, err := inst.GetPVCsUsingStorageClass(scname)
	if err != nil {
		return err
	}

	fmt.Printf("got pvcs: %v", pvcs)
	return nil
}

func demoBasic() error {
	inst := k8s.Instance()

	nodes, err := inst.GetNodes()
	if err != nil {
		return err
	}

	for _, n := range nodes.Items {
		fmt.Printf("node: %v\n", n)
	}

	return nil
}

func testPodsByNode(nodeName string) error {
	inst := k8s.Instance()

	pods, err := inst.GetPodsByNode(nodeName, "kube-system")
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		fmt.Printf("***** pod: %s (%s)", pod.Name, pod.Namespace)
	}
	return nil
}

func demoNs(kubeconfig string) error {
	client, err := loadClientFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}

	nsName := "dev"
	/*ns, err := client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: nsName,
		},
	})
	if err != nil {
		return err
	}*/

	ns, err := client.CoreV1().Namespaces().Get(nsName, meta_v1.GetOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("created ns: %s\n", ns.Name)

	deletePolicy := meta_v1.DeletePropagationOrphan
	err = client.CoreV1().Namespaces().Delete(nsName, &meta_v1.DeleteOptions{
		PropagationPolicy: &deletePolicy})
	if err != nil {
		return err
	}

	fmt.Printf("deleted ns: %s\n", ns.Name)
	return nil
}

func demoDocker() error {
	docker, err := dockerclient.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		return err
	}

	info, err := docker.Info()
	if err != nil {
		return err
	}

	fmt.Printf("docker security opts: %v\n", info.SecurityOptions)
	selinuxRegex := regexp.MustCompile("name=selinux")
	dockerSelinuxEnabled := false
	for _, opt := range info.SecurityOptions {
		if selinuxRegex.MatchString(opt) {
			dockerSelinuxEnabled = true
			break
		}
	}

	if !dockerSelinuxEnabled {
		fmt.Printf("selinux is disabled\n")
	} else {
		fmt.Printf("selinux is enabled\n")
	}

	return nil
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
