package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/portworx/sched-ops/k8s"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

func demoPXApps(scname, kubeconfig string) error {
	inst := k8s.Instance()
	/*pods, err := inst.GetPodsUsingVolumePluginByNodeName("k2n2", "kubernetes.io/portworx-volume")
	if err != nil {
		return err
	}

	logrus.Infof("found pods: %v", pods)*/

	deps, err := inst.GetDeploymentsUsingStorageClass(scname)
	if err != nil {
		return err
	}

	logrus.Infof("found deps: %v", deps)

	client, err := loadClientFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}

	for _, d := range deps {
		dCopy, err := client.Apps().Deployments(d.Namespace).Get(d.Name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		logrus.Infof("fetched: %v", dCopy)
	}

	ss, err := inst.GetStatefulSetsUsingStorageClass(scname)
	if err != nil {
		return err
	}

	logrus.Infof("found ss: %v", ss)
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

	logrus.Infof("got pvcs: %v", pvcs)
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

func main() {
	var kubeconfig string
	var node string
	var scname string

	//flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&kubeconfig, "kubeconfig", "", "- NOT RECOMMENDED FOR PRODUCTION - Path to kubeconfig.")
	flag.StringVar(&node, "node", "", "")
	flag.StringVar(&scname, "scname", "", "")
	flag.Parse()

	if len(kubeconfig) != 0 {
		fmt.Printf("using kubeconfig: %s\n", kubeconfig)
	} else {
		fmt.Printf("error: kubeconfig required\n")
		return
	}

	err := demoPVC(scname)
	if err != nil {
		fmt.Printf("PVC demo failed. err: %v\n", err)
		return
	}

	return
}
