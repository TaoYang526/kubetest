package common

import "gonum.org/v1/plot/vg"

const (
    // constants for testing environment
    QueueName        = "root.default"
    Namespace        = "default"
    HollowNodePrefix = "hollow-"

    // constants for yunikorn scheduler
    YSConfigMapNamespace            = "default"
    YSConfigMapName                 = "yunikorn-configs"
    YSConfigMapQueuesYamlKey        = "queues.yaml"
    YSConfigMapQueuesResourceMemKey = "memory"
    YSName                          = "yunikorn"

    // constants for ak8s scheduler
    K8SName = "default-scheduler"
    AK8SName = "ak8s-ee-scheduler"

    // constants for chart
    ChartWidth    = 6 * vg.Inch
    ChartHeight   = 6 * vg.Inch
    ChartSavePath = "/tmp/"
)

var (
    SchedulerNames = []string{YSName, K8SName, AK8SName}
    SchedulerAlias = map[string]string{YSName: YSName, K8SName: K8SName, AK8SName: AK8SName}
)