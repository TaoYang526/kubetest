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
    AppName = "throughput.test"
    ChartTitle = "Scheduling Throughput"
    ChartXLabel = "Seconds"
    ChartYLabel = "Number of Pods"
)

var (
    PodNum = 50000
    SelectPodLabels = map[string]string{cache.KeyApp: AppName}
)

func main() {
    // make sure all related pods are cleaned up
    monitor.WaitUtilAllMetricsAreCleanedUp(collectDeploymentMetrics)

    dataMap := make(map[string][]int, 2)
    for _, schedulerName := range common.SchedulerNames {
        fmt.Printf("Starting %s via scheduler %s\n", AppName, schedulerName)
        // create deployment
        deployment := cache.KubeDeployment{}.WithSchedulerName(schedulerName).WithAppName(
            AppName).WithPodNum(int32(PodNum)).Build()
        kubeclient.CreateDeployment(common.Namespace, deployment)
        beginTime := time.Now().Truncate(time.Second)
        // start monitor
        createMonitor := &monitor.Monitor{
            Name:           AppName + " create-monitor",
            Interval:       1,
            CollectMetrics: collectDeploymentMetrics,
            SkipSameMerics: true,
            StopTrigger: func(m *monitor.Monitor) bool {
                lastCp := m.GetLastCheckPoint()
                if lastCp.MetricValues[2] == PodNum {
                    // stop monitor when readyReplicas equals PodNum
                    return true
                }
                return false
            },
        }
        createMonitor.Start()
        // wait util this deployment is running successfully
        createMonitor.WaitForStopped()
        // calculate distribution of pod start times
        endTime := time.Now()
        podStartTimes := collector.CollectPodInfo(common.Namespace,
            kubeclient.GetListOptions(SelectPodLabels), collector.ParsePodStartTime)
        podStartTimeDistribution := collector.AnalyzeTimeDistribution(beginTime, endTime, podStartTimes)
        fmt.Printf("Distribution of pod start times: %v, seconds: %d beginTime: %v, endTime: %v \n",
            podStartTimeDistribution, len(podStartTimeDistribution), beginTime, endTime)

        // Save checkpoints
        dataMap[common.SchedulerAlias[schedulerName]] = podStartTimeDistribution

        // delete deployment
        kubeclient.DeleteDeployment(common.Namespace, AppName)
        // make sure all related pods are cleaned up
        monitor.WaitUtilAllMetricsAreCleanedUp(collectDeploymentMetrics)
    }

    // draw chart
    linePoints := painter.GetLinePoints(dataMap)
    chart := &painter.Chart{
        Title:      ChartTitle,
        XLabel:     ChartXLabel,
        YLabel:     ChartYLabel,
        Width:      common.ChartWidth,
        Height:     common.ChartHeight,
        LinePoints: linePoints,
        SvgFile:    common.ChartSavePath + "throughput.svg",
    }
    painter.DrawChart(chart)
}

func collectDeploymentMetrics() []int {
    return collector.CollectDeploymentMetrics(common.Namespace, AppName)
}