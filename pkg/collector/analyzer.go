package collector

import (
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/common"
    "math"
    "time"
)

func AnalyzeTimeDistribution(beginTime time.Time, endTime time.Time, ifSlice []interface{}) []int {
    maxSeconds := int(math.Ceil(endTime.Sub(beginTime).Seconds()))
    distribution := make([]int, maxSeconds+1)
    maxFlag := 0
    for _, ifItem := range ifSlice {
        if columns,ok := ifItem.([]interface{}); ok {
            podStartTime := common.ConvertToTime(beginTime, endTime, columns[0])
            seconds := int(math.Ceil(podStartTime.Sub(beginTime).Seconds()))
            distribution[seconds] += 1
            if seconds > maxFlag {
                maxFlag = seconds
                fmt.Printf("-------->createTime:%v, startTime:%v, podName:%v, nodeName:%v\n", columns[1], columns[0], columns[2], columns[3])
            }
        } else {
            panic(fmt.Errorf("Error: type of data items is not []interface{}, data: %v ", ifItem))
        }
    }
    fmt.Printf("------>maxSeconds: %d, distribution:%v", maxSeconds, distribution)
    lastNonZeroIndex := len(distribution) -1
    for i := len(distribution) - 1; i >= 0; i-- {
        if distribution[i] != 0 {
            lastNonZeroIndex = i
            break
        }
    }
    distribution = distribution[:lastNonZeroIndex+1]
    cumulativeNum := 0
    for i := range distribution {
        cumulativeNum += distribution[i]
        distribution[i] = cumulativeNum
    }
    return distribution
}

func AnalyzeResourceDistribution(beginTime time.Time, endTime time.Time, ifSlice []interface{},
    allocatableResources map[string]int64) [10][]int {
    // prepare allocatable&allocated resources
    resources := map[string][]int64{}
    for additionalInfo, allocatableResourceValue := range allocatableResources {
        resources[additionalInfo] = []int64{allocatableResourceValue, 0}
    }
    maxSeconds := int(math.Ceil(endTime.Sub(beginTime).Seconds()))
    // parse pod infos
    sortedPodInfos := make([][][]interface{}, maxSeconds)
    maxFlag := 0
    for _, ifItem := range ifSlice {
        if columns,ok := ifItem.([]interface{}); ok {
            podStartTime := common.ConvertToTime(beginTime, endTime, columns[0])
            seconds := int(math.Ceil(podStartTime.Sub(beginTime).Seconds()))
            if seconds > maxFlag {
                maxFlag = seconds
                fmt.Printf("-------->createTime:%v, startTime:%v, podName:%v\n", columns[1], columns[0], columns[2])
            }
            sortedPodInfos[seconds] = append(sortedPodInfos[seconds], columns)
        } else {
            panic(fmt.Errorf("Error: type of data items is not []interface{}, data: %v ", columns))
        }
    }

    // add distribution into buckets in chronological order
    var buckets [10][]int
    lastSeconds := 0
    for seconds, aggregatedIfItems := range sortedPodInfos {
        if aggregatedIfItems == nil {
            continue
        }
        for _, columns := range aggregatedIfItems {
            podResourceMilliValue := common.ConvertToInt64(columns[1])
            additionalInfo := common.ConvertToString(columns[2])
            if seconds != lastSeconds {
                // calculate distribution of resources and set every second from lastSeconds to seconds-1
                resourceDistribution := calculateResourceDistribution(resources)
                for lastSeconds < seconds {
                    for i,v := range resourceDistribution {
                        buckets[i] = append(buckets[i], v)
                    }
                    lastSeconds++
                }
            }
            // update node resource
            if _, ok := resources[additionalInfo]; !ok {
                panic(fmt.Errorf("Error additional info: %s, resources: %+v ", additionalInfo, resources))
            }
            resources[additionalInfo][1] += podResourceMilliValue
        }
    }
    return buckets
}

func calculateResourceDistribution(resources map[string][]int64) [10]int {
    var buckets [10]int
    for _, resourceValues := range resources {
        if len(resourceValues) == 2 {
            ratio := float32(resourceValues[1]*10) / float32(resourceValues[0])
            bucketIndex := int(ratio)
            if bucketIndex == 10 {
                bucketIndex = 9
            }
            buckets[bucketIndex] += 1
        } else {
            panic(fmt.Errorf("Error resources: %+v ", resourceValues))
        }
    }
    return buckets
}


func AnalyzeUsageRatioDistribution(beginTime time.Time, endTime time.Time, ifSlice []interface{},
    allocatableResources map[string]int64) map[string][]float64 {
    // prepare allocatable&allocated resources
    resources := map[string][]int64{}
    resourceUsageRatios := map[string][]float64{}
    for additionalInfo, allocatableResourceValue := range allocatableResources {
        resources[additionalInfo] = []int64{allocatableResourceValue, 0}
        resourceUsageRatios[additionalInfo] = []float64{}
    }
    maxSeconds := int(math.Ceil(endTime.Sub(beginTime).Seconds()))
    // parse pod infos
    sortedPodInfos := make([][][]interface{}, maxSeconds)
    for _, ifItem := range ifSlice {
        if columns,ok := ifItem.([]interface{}); ok {
            podStartTime := common.ConvertToTime(beginTime, endTime, columns[0])
            seconds := int(math.Ceil(podStartTime.Sub(beginTime).Seconds()))
            sortedPodInfos[seconds] = append(sortedPodInfos[seconds], columns)
        } else {
            panic(fmt.Errorf("Error: type of data items is not []interface{}, data: %v ", columns))
        }
    }

    // add distribution into buckets in chronological order
    lastSeconds := 0
    for seconds, aggregatedIfItems := range sortedPodInfos {
        if aggregatedIfItems == nil {
            continue
        }
        for _, columns := range aggregatedIfItems {
            podResourceMilliValue := common.ConvertToInt64(columns[1])
            additionalInfo := common.ConvertToString(columns[2])
            if seconds != lastSeconds {
                // calculate distribution of resources and set every second from lastSeconds to seconds-1
                resourceDistribution := calculateUsageRatioDistribution(resources)
                for lastSeconds < seconds {
                    for resourceName,v := range resourceDistribution {
                        resourceUsageRatios[resourceName] = append(resourceUsageRatios[resourceName], v)
                    }
                    lastSeconds++
                }
            }
            // update node resource
            resources[additionalInfo][1] += podResourceMilliValue
        }
    }
    return resourceUsageRatios
}

func calculateUsageRatioDistribution(resources map[string][]int64) map[string]float64 {
    usageRatios := map[string]float64{}
    for name, resourceValues := range resources {
        if len(resourceValues) == 2 {
            usageRatios[name] = float64(resourceValues[1]) / float64(resourceValues[0])
        } else {
            panic(fmt.Errorf("Error resources: %+v ", resourceValues))
        }
    }
    return usageRatios
}

