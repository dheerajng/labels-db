package main

import (
	"labels-db/client"
	"os"
	"strings"
	"time"

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
	// Wait for sometime for redis DB to come up
	time.Sleep(10 * time.Second)
	// Create Labels DB
	if err = ldbClient.CreateLablesDB(); err != nil {
		logrus.Errorf("Could not retrieve Service details")
		return
	}

	// Create a HTTP Server
	client.CreateLabelsServer()
}
