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

	greeter_api "github.com/imind-lab/greeter-api/server/proto/greeter-api"
	"github.com/imind-lab/greeter-api/server/service"
	"github.com/imind-lab/micro"
	"github.com/imind-lab/micro/grpcx"
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
