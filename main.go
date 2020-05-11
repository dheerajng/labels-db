package main

import (
	"labels-db/client"
)

func main() {
	var k8sConfig []byte
	// creates the in-cluster config
	ldbClient, err := client.NewK8sClient(k8sConfig, "")
	if err != nil {
		panic(err.Error())
	}
	if err := ldbClient.GetPodsLabels(); err != nil {
		panic(err.Error())
	}
}
