/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package server

import (
	"fmt"
	_ "net/http/pprof"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/imind-lab/greeter-api/application/greeter-api/proto"
	"github.com/imind-lab/greeter-api/application/greeter-api/service"
	"github.com/imind-lab/micro"
	grpcx "github.com/imind-lab/micro/grpc"
)

func Serve() error {
	svc := micro.NewService()

	grpcCred := grpcx.NewGrpcCred()

	svc.Init(
		micro.ServerCred(grpcCred.ServerCred()),
		micro.ClientCred(grpcCred.ClientCred()))

	grpcSrv := svc.GrpcServer()
	greeter_api.RegisterGreeterServiceServer(grpcSrv, service.NewGreeterService(svc.Options().Logger))

	// 注册gRPC-Gateway
	endPoint := fmt.Sprintf(":%d", viper.GetInt("service.port.grpc"))

	mux := svc.ServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(grpcCred.ClientCred())}
	err := greeter_api.RegisterGreeterServiceHandlerFromEndpoint(svc.Options().Context, mux, endPoint, opts)
	if err != nil {
		return err
	}
	return svc.Run()
}
