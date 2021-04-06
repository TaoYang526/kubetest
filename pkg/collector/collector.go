package collector

import (
    "context"
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/cache"
    "github.com/TaoYang526/kubetest/pkg/common"
    "github.com/TaoYang526/kubetest/pkg/kubeclient"
    "github.com/apache/incubator-yunikorn-core/pkg/common/configs"
    "gopkg.in/yaml.v2"
    apiv1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "time"
)

func CollectPodInfo(namespace string, listOptions *metav1.ListOptions,
    parsePodInfoFunc func(*apiv1.Pod) []interface{}) []interface{} {
    pods, err := kubeclient.GetPods(namespace, listOptions)
    if err != nil {
        fmt.Printf("List pods failed: %v\n", err)
    }
    if pods != nil {
        var podInfos []interface{}
        for _, pod := range pods.Items {
            podInfo := parsePodInfoFunc(&pod)
            if podInfo == nil {
                continue
            }
            podInfos = append(podInfos, podInfo)
        }
        return podInfos
    } else {
        fmt.Println("Get nil pods")
        return nil
    }
}

// metrics: [numCreatedPods, numScheduledPods, numRunningPods]
func CollectPodMetrics(namespace string, listOptions *metav1.ListOptions) []int {
    pods, err := kubeclient.GetPods(namespace, listOptions)
    if err != nil {
        fmt.Printf("List pods failed: %v\n", err)
    }
    if pods != nil {
        numCreated := len(pods.Items)
        numScheduled := 0
        numRunning := 0
        for _, pod := range pods.Items {
            if pod.Status.Phase == "Running" {
                numRunning += 1
                numScheduled += 1
            } else if pod.Status.HostIP != "" {
                numScheduled += 1
            }
        }
        var metrics []int
        metrics = append(metrics, numCreated)
        metrics = append(metrics, numScheduled)
        metrics = append(metrics, numRunning)
        curTime := time.Now()
        fmt.Printf("%d:%d:%d created/scheduled/running pods: %d/%d/%d \n",
           curTime.Hour(), curTime.Minute(), curTime.Second(),
           numCreated, numScheduled, numRunning)
        return metrics
    } else {
        fmt.Println("Get nil pods")
        return []int{0, 0, 0}
    }
}

// replicas metrics of deployment: [desired, created, ready]
func CollectDeploymentMetrics(namespace string, appName string) []int {
    deploymentsClient := kubeclient.GetClientSet().AppsV1().Deployments(namespace)
    deployment, err := deploymentsClient.Get(context.TODO(), appName, metav1.GetOptions{})
    if deployment == nil || err != nil {
        fmt.Println("Failed to get deployment: ", err.Error())
        return []int{0, 0, 0}
    }
    curTime := time.Now()
    fmt.Printf("%d:%d:%d desired/created/ready replicas: %d/%d/%d \n",
        curTime.Hour(), curTime.Minute(), curTime.Second(),
        *deployment.Spec.Replicas, deployment.Status.Replicas, deployment.Status.ReadyReplicas)
    return []int{int(*deployment.Spec.Replicas), int(deployment.Status.Replicas), int(deployment.Status.ReadyReplicas)}
}

func CollectQueueInfo() (*configs.SchedulerConfig, error) {
    configMap, err := kubeclient.GetConfigMap(
        common.YSConfigMapNamespace, common.YSConfigMapName, &metav1.GetOptions{})
    if err != nil {
        fmt.Printf("Get configmap failed: %v\n", err)
        return nil, err
    }
    content := configMap.Data[common.YSConfigMapQueuesYamlKey]
    schedulerConfig := configs.SchedulerConfig{}
    yaml.Unmarshal([]byte(content), &schedulerConfig)
    if err != nil {
        fmt.Printf("failed to parse queue configuration: %v\n", err)
        return nil, err
    }
    return &schedulerConfig, nil
}

// deprecated time-consuming function: above 10s for 5k+ pods
func CollectNodeMetrics(nodeAllocatableResources map[string]int64,
    namespace string, podListOptions *metav1.ListOptions) []int {
    // prepare node allocatable&allocated resources
    nodeResources := map[string][]int64{}
    for nodeName, allocatableResourceValue := range nodeAllocatableResources {
        nodeResources[nodeName] = []int64{allocatableResourceValue, 0}
    }
    // get pods & update node metrics
    pods, _ := kubeclient.GetPods(namespace, podListOptions)
    if pods != nil {
        for _, pod := range pods.Items {
            nodeName := pod.Spec.NodeName
            if _, ok := nodeResources[nodeName]; ok {
                var allocatedResource int64
                if mem, ok := pod.Labels[cache.KeyMemRequest]; ok {
                    quantity := resource.MustParse(mem)
                    allocatedResource += quantity.MilliValue()
                }
                nodeResources[nodeName][1] += allocatedResource
            }
        }
    } else {
        fmt.Println("Get nil nodes")
    }
    // calculate node metrics
    buckets := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    for nodeName, nodeResourceValues := range nodeResources {
        if len(nodeResourceValues) == 2 {
            ratio := float32(nodeResourceValues[1] * 10) / float32(nodeResourceValues[0])
            bucketIndex := int(ratio)
            buckets[bucketIndex] += 1
        } else {
            fmt.Println("No ", nodeName, len(nodeResourceValues))
        }
    }
    return buckets
}