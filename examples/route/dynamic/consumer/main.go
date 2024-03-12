/**
 * Tencent is pleased to support the open source community by making polaris-go available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/pkg/model"
	"log"
	"net"
	"strings"
)

var (
	namespace     string
	service       string
	selfNamespace string
	selfService   string
	port          int64
	token         string = "xxx"
)

func initArgs() {
	flag.StringVar(&namespace, "namespace", "default", "namespace")
	flag.StringVar(&service, "service", "RouteEchoServer", "service")
	flag.StringVar(&selfNamespace, "selfNamespace", "default", "selfNamespace")
	flag.StringVar(&selfService, "selfService", "", "selfService")
	flag.Int64Var(&port, "port", 18080, "port")
	flag.StringVar(&token, "token", "FPI+K9USIvHYU8JUljM3TqAg1Wizxta7i+WEi73RkDMQl1HhIBoIc+EKYinqiViTx7TJlBJSY2/R/tXfZkGv8mGB", "token")
}

// PolarisConsumer .
type PolarisConsumer struct {
	consumer  polaris.ConsumerAPI
	router    polaris.RouterAPI
	namespace string
	service   string
}

// Run .
func (svr *PolarisConsumer) Run() {
	getAllRequest := &polaris.GetOneInstanceRequest{}
	println(namespace, service, token)
	getAllRequest.Namespace = namespace
	getAllRequest.Service = service
	getAllRequest.Token = token
	oneInstResp, err := svr.consumer.GetOneInstance(getAllRequest)
	if nil != err {
		log.Printf("[error] fail to GetOneInstance, err is %v", err)
	} else {
		p := oneInstResp.GetInstance().GetPort()
		fmt.Printf("port: %d", p)
	}

	//_ = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	return
}

func main() {
	initArgs()
	flag.Parse()
	if len(namespace) == 0 || len(service) == 0 {
		log.Print("namespace and service are required")
		return
	}
	sdkCtx, err := polaris.NewSDKContextByAddress("localhost:8091")
	if nil != err {
		log.Fatalf("fail to create sdk context, err is %v", err)
	}
	defer sdkCtx.Destroy()

	svr := &PolarisConsumer{
		consumer:  polaris.NewConsumerAPIByContext(sdkCtx),
		router:    polaris.NewRouterAPIByContext(sdkCtx),
		namespace: namespace,
		service:   service,
	}

	svr.Run()

}

func convertQuery(rawQuery string) []model.Argument {
	arguments := make([]model.Argument, 0, 4)
	if len(rawQuery) == 0 {
		return arguments
	}
	tokens := strings.Split(rawQuery, "&")
	if len(tokens) > 0 {
		for _, token := range tokens {
			values := strings.Split(token, "=")
			arguments = append(arguments, model.BuildQueryArgument(values[0], values[1]))
		}
	}
	return arguments
}

func getLocalHost(serverAddr string) (string, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if nil != err {
		return "", err
	}
	localAddr := conn.LocalAddr().String()
	colonIdx := strings.LastIndex(localAddr, ":")
	if colonIdx > 0 {
		return localAddr[:colonIdx], nil
	}
	return localAddr, nil
}
