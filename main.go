package main

import (
	"labels-db/client"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func main() {
	var k8sConfig []byte
	if strings.EqualFold(os.Getenv("DEBUG"), "true") == true {
		logrus.SetLevel(logrus.DebugLevel)
	}
	// creates the in-cluster config
	ldbClient, err := client.NewK8sClient(k8sConfig, "")
	if err != nil {
		panic(err.Error())
	}
	// List down Pods' labels
	if err = ldbClient.GetPodsLabels(); err != nil {
		panic(err.Error())
	}
	if err = ldbClient.GetSvcDetails(); err != nil {
		logrus.Errorf("Could not retrieve Service details")
	}
}
