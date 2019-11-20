package main

import (
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/cache"
    "github.com/TaoYang526/kubetest/pkg/collector"
    "github.com/TaoYang526/kubetest/pkg/common"
    "github.com/TaoYang526/kubetest/pkg/kubeclient"
    "github.com/TaoYang526/kubetest/pkg/monitor"
    "github.com/TaoYang526/kubetest/pkg/painter"
    "time"
)

const (
    AppNamePrefix = "queue-fairness.test."
    ChartTitle = "Queue Fairness"
    ChartXLabel = "Seconds"
    ChartYLabel = "Usage Ratio (=AllocatedResource/GuaranteedResource)"
)

var (
    SelectPodLabels = map[string]string{cache.KeyGroup: cache.DefaultGroupName}
    FieldSelect = map[string]string{}
    TestApps    = map[string][]interface{}{
        "app1": {"root.default", 1000, "10Mi", "10m"},
        "app2": {"root.test", 500, "10Mi", "10m"},
        "app3": {"root.sandbox", 200, "10Mi", "10m"},
        "app4": {"root.search", 200, "50Mi", "50m"},
    }
)

func main() {
    // make sure all related pods are cleaned up
    monitor.WaitUtilAllMetricsAreCleanedUp(collectPodMetrics)

    // prepare queue allocatable resources
    config, _ := collector.CollectQueueInfo()
    queueGuaranteedResources := collector.ParseQueueGuaranteedResources(config)
    fmt.Printf("Got guaranteed resources from %d queues: %+v\n",
        len(queueGuaranteedResources), queueGuaranteedResources)

    // create deployments among queues
    beginTime := time.Now().Truncate(time.Second).Add(-1 * time.Second)
    fmt.Printf("Starting queue fairness test via scheduler %s, begin time: %v \n",
        common.YSName, beginTime)

    // create deployment
    totalPodNum := 0
    for appName,appConfigs := range TestApps {
        queuePath := common.ConvertToString(appConfigs[0])
        podNum := common.ConvertToInt(appConfigs[1])
        totalPodNum += podNum
        podMem := common.ConvertToString(appConfigs[2])
        podCPU := common.ConvertToString(appConfigs[3])
        deployment := cache.KubeDeployment{}.WithSchedulerName(common.YSName).WithQueuePath(
            queuePath).WithAppName(AppNamePrefix + appName).WithPodNum(int32(podNum)).WithResourceMemLimit(
            podMem).WithResourceMemRequest(podMem).WithResourceCPULimit(podCPU).WithResourceCPURequest(podCPU).Build()
        kubeclient.CreateDeployment(common.Namespace, deployment)
    }

    // start monitor
    createMonitor := &monitor.Monitor{
        Name:        AppNamePrefix + " create-monitor",
        Interval:    1,
        CollectMetrics: collectPodMetrics,
        StopTrigger: func (m *monitor.Monitor) bool {
            lastCp := m.GetLastCheckPoint()
            if lastCp.MetricValues[1] == totalPodNum {
                return true
            }
            return false
        },
    }
    createMonitor.Start()
    // wait util this deployment is running successfully
    createMonitor.WaitForStopped()
    endTime := time.Now()
    podResourceInfos := collector.CollectPodInfo(common.Namespace,
        kubeclient.GetListOptions(SelectPodLabels), collector.ParsePodResourceInfoWithQueueName)
    queueResourceDistributions := collector.AnalyzeUsageRatioDistribution(beginTime, endTime,
        podResourceInfos, queueGuaranteedResources)

    // prepare line points
    var linePoints []interface{}
    for queuePath,usageRatioDistribution := range queueResourceDistributions {
        linePoints = append(linePoints, queuePath, painter.GetPointsFromFloat64Slice(usageRatioDistribution))
    }
    // draw chart
    chart := &painter.Chart{
        Title:      ChartTitle,
        XLabel:     ChartXLabel,
        YLabel:     ChartYLabel,
        Width:      common.ChartWidth,
        Height:     common.ChartHeight,
        SvgFile:    common.ChartSavePath + "queue-fairness-" + common.SchedulerAlias[common.YSName] + ".svg",
        LinePoints: linePoints,
    }
    painter.DrawChart(chart)

    // delete deployment
    for appName := range TestApps {
        kubeclient.DeleteDeployment(common.Namespace, AppNamePrefix + appName)
    }

    deleteMonitor := &monitor.Monitor{
        Name:        AppNamePrefix + " delete-monitor",
        Interval:    5,
        CollectMetrics: collectPodMetrics,
        StopTrigger: func (m *monitor.Monitor) bool {
            lastCp := m.GetLastCheckPoint()
            if lastCp.MetricValues[0] == 0 {
                return true
            }
            return false
        },
    }
    deleteMonitor.Start()
    deleteMonitor.WaitForStopped()
}

func collectPodMetrics() []int {
    return collector.CollectPodMetrics(common.Namespace, kubeclient.GetListOptions(SelectPodLabels))
}