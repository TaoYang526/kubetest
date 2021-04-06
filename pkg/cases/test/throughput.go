package main

import (
    "github.com/TaoYang526/kubetest/pkg/kubeclient"
)

func main() {
    // get deployment
    deployment := kubeclient.GetDeployment("default", "yunikorn-scheduler")

    // delete deployment
    kubeclient.DeleteDeployment("default", "yunikorn-scheduler")
}

