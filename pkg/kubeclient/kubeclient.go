package kubeclient

import (
    "context"
    "flag"
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/common"
    appsv1 "k8s.io/api/apps/v1"
    apiv1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
    "path/filepath"
    "strings"
)

var clientSet *kubernetes.Clientset

func init() {
    var kubeconfig *string
    if home := homedir.HomeDir(); home != "" {
        fmt.Sprintf("Home dir: %s", home)
        kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
    } else {
        fmt.Println("Env HOME not defined!")
        kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
    }
    flag.Parse()
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
    if err != nil {
        panic(err)
    }
    clientSet, err = kubernetes.NewForConfig(config)
    if err != nil {
        panic(err)
    }
}

func GetListOptions(selectLabels map[string]string) *metav1.ListOptions{
    labelSelector := metav1.LabelSelector{MatchLabels: selectLabels}
    return &metav1.ListOptions{
        LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
    }
}

func GetPods(namespace string, listOptions *metav1.ListOptions) (*apiv1.PodList, error) {
    return clientSet.CoreV1().Pods(namespace).List(context.TODO(), *listOptions)
}

func GetNodes(listOptions *metav1.ListOptions) (*apiv1.NodeList, error) {
    return clientSet.CoreV1().Nodes().List(context.TODO(), *listOptions)
}

func GetConfigMap(namespace string, name string, getOptions *metav1.GetOptions) (*apiv1.ConfigMap, error) {
    return clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, *getOptions)
}

func GetClientSet() *kubernetes.Clientset {
    return clientSet
}

// Get hollow node resources: key - nodeName, value - allocatable resource
func GetHollowNodeAllocatableResources() map[string]int64 {
    nodes, err := GetNodes(&metav1.ListOptions{LabelSelector: labels.Everything().String()})
    if err != nil {
        fmt.Printf("List nodes failed: %v\n", err)
        return nil
    }
    if nodes == nil {
        fmt.Println("Get nil nodes")
        return nil
    }
    nodeResources := map[string]int64{}
    for _, node := range nodes.Items {
        if strings.HasPrefix(node.Name, common.HollowNodePrefix) {
            nodeResources[node.Name] = node.Status.Allocatable.Memory().MilliValue()
        }
    }
    return nodeResources
}

func CreateDeployment(namespace string, deployment *appsv1.Deployment) {
    deploymentsClient := clientSet.AppsV1().Deployments(namespace)
    fmt.Println("Creating deployment...")
    result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func GetDeployment(namespace, appName string) *appsv1.Deployment {
    deploymentsClient := clientSet.AppsV1().Deployments(namespace)
    if deployment, err := deploymentsClient.Get(context.TODO(), appName, metav1.GetOptions{}); err!=nil {
        panic(err)
    } else {
        return deployment
    }
}

func DeleteDeployment(namespace, appName string) {
    deploymentsClient := clientSet.AppsV1().Deployments(namespace)
    deletePolicy := metav1.DeletePropagationForeground
    if err := deploymentsClient.Delete(context.TODO(), appName, metav1.DeleteOptions{
        PropagationPolicy: &deletePolicy,
    }); err != nil {
        panic(err)
    }
    fmt.Println("Deleted deployment.")
}