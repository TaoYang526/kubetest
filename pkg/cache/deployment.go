package cache

import (
    "fmt"
    "github.com/TaoYang526/kubetest/pkg/common"
    appsv1 "k8s.io/api/apps/v1"
    apiv1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
    KeyGroup = "group"
    KeyApp = "app"
    KeyAppID = "applicationId"
    KeyQueue = "queue"
    KeyMemRequest = "memRequest"
    DefaultAppName = "kubetest-app"
    DefaultGroupName = "kubetest-group"
    DefaultQueueName = common.QueueName
    DefaultPodNum = int32(10)
    DefaultContainerName = "web"
    DefaultContainerImage = "nginx:1.12"
    DefaultContainerPortName = "http"
    DefaultSchedulerName = ""
    DefaultResourceCPULimit = "20m"
    DefaultResourceMemLimit = "20Mi"
    DefaultResourceCPURequest = "10m"
    DefaultResourceMemRequest = "10Mi"
)

type KubeDeployment struct {
    groupName          string
    schedulerName      string
    queuePath          string
    appName            string
    podNum             int32
    resourceCPURequest string
    resourceMemRequest string
    resourceCPULimit   string
    resourceMemLimit   string
}

func (d KubeDeployment) WithAppName(appName string) KubeDeployment {
    d.appName = appName
    return d
}

func (d KubeDeployment) WithQueuePath(queuePath string) KubeDeployment {
    d.queuePath = queuePath
    return d
}

func (d KubeDeployment) WithGroupName(groupName string) KubeDeployment {
    d.groupName = groupName
    return d
}

func (d KubeDeployment) WithResourceCPURequest(resourceCPURequest string) KubeDeployment {
    d.resourceCPURequest = resourceCPURequest
    return d
}

func (d KubeDeployment) WithResourceMemRequest(resourceMemRequest string) KubeDeployment {
    d.resourceMemRequest = resourceMemRequest
    return d
}

func (d KubeDeployment) WithResourceCPULimit(resourceCPULimit string) KubeDeployment {
    d.resourceCPULimit = resourceCPULimit
    return d
}

func (d KubeDeployment) WithResourceMemLimit(resourceMemLimit string) KubeDeployment {
    d.resourceMemLimit = resourceMemLimit
    return d
}

func (d KubeDeployment) WithPodNum(podNum int32) KubeDeployment {
    d.podNum = podNum
    return d
}

func (d KubeDeployment) WithSchedulerName(schedulerName string) KubeDeployment {
    d.schedulerName = schedulerName
    return d
}

func (d KubeDeployment) Build() *appsv1.Deployment {
    if d.groupName == "" {
        d.groupName = DefaultGroupName
    }
    if d.schedulerName == "" {
        d.schedulerName = DefaultSchedulerName
    }
    if d.appName == "" {
        d.appName = DefaultAppName
    }
    if d.queuePath == "" {
        d.queuePath = DefaultQueueName
    }
    if d.podNum == 0 {
        d.podNum = DefaultPodNum
    }
    if d.resourceCPULimit == "" {
        d.resourceCPULimit = DefaultResourceCPULimit
    }
    if d.resourceMemLimit == "" {
        d.resourceMemLimit = DefaultResourceMemLimit
    }
    if d.resourceCPURequest == "" {
        d.resourceCPURequest = DefaultResourceCPURequest
    }
    if d.resourceMemRequest == "" {
        d.resourceMemRequest = DefaultResourceMemRequest
    }
    fmt.Printf("Build KubeDeployment: %+v\n", d)
    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: d.appName,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &d.podNum,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    KeyApp: d.appName,
                },
            },
            Template: apiv1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        KeyGroup: d.groupName,
                        KeyApp:   d.appName,
                        KeyAppID: d.appName,
                        KeyQueue: d.queuePath,
                        KeyMemRequest: d.resourceMemRequest,
                    },
                },
                Spec: apiv1.PodSpec{
                    SchedulerName: d.schedulerName,
                    Containers: []apiv1.Container{
                        {
                            Name:  DefaultContainerName,
                            Image: DefaultContainerImage,
                            Ports: []apiv1.ContainerPort{
                                {
                                    Name:          DefaultContainerPortName,
                                    Protocol:      apiv1.ProtocolTCP,
                                    ContainerPort: 80,
                                },
                            },
                            Resources: apiv1.ResourceRequirements {
                                Limits:   apiv1.ResourceList{
                                    apiv1.ResourceCPU: resource.MustParse(d.resourceCPULimit),
                                    apiv1.ResourceMemory: resource.MustParse(d.resourceMemLimit),
                                },
                                Requests: apiv1.ResourceList{
                                    apiv1.ResourceCPU: resource.MustParse(d.resourceCPURequest),
                                    apiv1.ResourceMemory: resource.MustParse(d.resourceMemRequest),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}