package collector

import (
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/cache"
    "github.com/TaoYang526/kubetest/pkg/common"
    "github.com/cloudera/yunikorn-core/pkg/common/configs"
    apiv1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
)

// columns of result item: podStartTime
func ParsePodStartTime(pod *apiv1.Pod) []interface{} {
    if pod.Status.StartTime == nil {
        fmt.Printf("Pod without start time: %v\n", pod)
        return nil
    }
    //fmt.Printf("----->node: %v, startTime: %v \n", pod.Name, pod.Status.StartTime)
    return []interface{}{pod.Status.StartTime.Time}
}

// columns of result item: podStartTime, resourceValue, additionalInfo
func ParsePodResourceInfo(pod *apiv1.Pod, additionalInfo string) []interface{} {
    if pod.Status.StartTime == nil || additionalInfo == "" {
        panic(fmt.Errorf("Parse resource info failed for pod without start time or additional info: %v \n", pod))
        return nil
    }
    var memRequestMilliValue int64
    if mem, ok := pod.Labels[cache.KeyMemRequest]; ok {
        quantity := resource.MustParse(mem)
        memRequestMilliValue = quantity.MilliValue()
    } else {
        panic(fmt.Errorf("Parse resource info failed for pod: %v\n", pod))
    }
    return []interface{}{pod.Status.StartTime.Time, memRequestMilliValue, additionalInfo}
}

func ParsePodResourceInfoWithNodeName(pod *apiv1.Pod) []interface{} {
    return ParsePodResourceInfo(pod, pod.Spec.NodeName)
}

func ParsePodResourceInfoWithQueueName(pod *apiv1.Pod) []interface{} {
    if queueName, ok := pod.Labels[cache.KeyQueue]; ok {
        return ParsePodResourceInfo(pod, queueName)
    } else {
        panic(fmt.Errorf("Parse resource info failed for pod: %v\n", pod))
    }
    return nil
}

func ParseQueueGuaranteedResources(config *configs.SchedulerConfig) map[string]int64 {
    queueResources := map[string]int64{}
    for _, partition := range config.Partitions {
        for _, queueConfig := range partition.Queues {
            parseQueueResource("", queueConfig, queueResources)
        }
    }
    return queueResources
}

func parseQueueResource(queuePrefix string, queueConfig configs.QueueConfig, queueResources map[string]int64) {
    if len(queuePrefix) > 0 {
        queuePrefix += "."
    }
    if len(queueConfig.Queues) > 0 {
        for _, subQueueConfig := range queueConfig.Queues {
            parseQueueResource(queuePrefix+queueConfig.Name, subQueueConfig, queueResources)
        }
    } else {
        queueGuaranteedMem := queueConfig.Resources.Guaranteed[common.YSConfigMapQueuesResourceMemKey]
        quantity := resource.MustParse(queueGuaranteedMem+"Mi")
        queueResources[queuePrefix+queueConfig.Name] = quantity.MilliValue()
    }
}
