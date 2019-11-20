package main

import (
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/cache"
    "github.com/TaoYang526/kubetest/pkg/collector"
    "github.com/TaoYang526/kubetest/pkg/common"
    "github.com/TaoYang526/kubetest/pkg/kubeclient"
    "github.com/TaoYang526/kubetest/pkg/monitor"
    "github.com/TaoYang526/kubetest/pkg/painter"
    "strconv"
    "time"
)

const (
    AppName = "node-fairness.test"
    ChartTitle = "Node Fairness"
    ChartXLabel = "Seconds"
    ChartYLabel = "Number of Nodes"
)

var (
    PodMem = "190Mi"
    PodNum = 40000
    SelectPodLabels = map[string]string{cache.KeyApp: AppName}
    FieldSelect = map[string]string{}
    nodeAllocatableResources = kubeclient.GetHollowNodeAllocatableResources()
)

func main() {
    fmt.Printf("Got allcatable resources from %d hollow nodes\n", len(nodeAllocatableResources))

    // make sure all related pods are cleaned up
    monitor.WaitUtilAllMetricsAreCleanedUp(collectDeploymentMetrics)

    // prepare node allocatable resources
    for _, schedulerName := range common.SchedulerNames {
        beginTime := time.Now().Truncate(time.Second).Add(-1 * time.Second)
        fmt.Printf("Starting %s via scheduler %s, begin time: %v \n", AppName, schedulerName, beginTime)

        // create deployment
        deployment := cache.KubeDeployment{}.WithSchedulerName(schedulerName).WithAppName(
            AppName).WithPodNum(int32(PodNum)).WithResourceMemLimit(PodMem).WithResourceMemRequest(PodMem).Build()
        kubeclient.CreateDeployment(common.Namespace, deployment)
        // start monitor
        createMonitor := &monitor.Monitor{
            Name:        AppName + " create monitor",
            Interval:    1,
            CollectMetrics: collectDeploymentMetrics,
            StopTrigger: func (m *monitor.Monitor) bool {
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
        endTime := time.Now()
        podResourceInfos := collector.CollectPodInfo(common.Namespace,
            kubeclient.GetListOptions(SelectPodLabels), collector.ParsePodResourceInfoWithNodeName)
        nodeResourceDistributions := collector.AnalyzeResourceDistribution(beginTime, endTime,
            podResourceInfos, nodeAllocatableResources)

        // prepare line points
        var linePoints []interface{}
        for i,v := range nodeResourceDistributions {
            typeName := "bucket-" + strconv.Itoa(i)
            linePoints = append(linePoints, typeName, painter.GetPointsFromSlice(v))
        }
        // draw chart
        chart := &painter.Chart{
            Title:      ChartTitle,
            XLabel:     ChartXLabel,
            YLabel:     ChartYLabel,
            Width:      common.ChartWidth,
            Height:     common.ChartHeight,
            LinePoints: linePoints,
            SvgFile:    common.ChartSavePath + "node-fairness-" + common.SchedulerAlias[schedulerName] + ".svg",
        }
        painter.DrawChart(chart)

        // delete deployment
        kubeclient.DeleteDeployment(common.Namespace, AppName)
        // make sure all related pods are cleaned up
        monitor.WaitUtilAllMetricsAreCleanedUp(collectDeploymentMetrics)
    }
}

func collectDeploymentMetrics() []int {
    return collector.CollectDeploymentMetrics(common.Namespace, AppName)
}