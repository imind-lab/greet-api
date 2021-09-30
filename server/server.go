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

	greet_api "github.com/imind-lab/greet-api/server/proto/greet-api"
	"github.com/imind-lab/greet-api/server/service"
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
	greet_api.RegisterGreetServiceServer(grpcSrv, service.NewGreetService(svc.Options().Logger))

	// 注册gRPC-Gateway
	endPoint := fmt.Sprintf(":%d", viper.GetInt("service.port.grpc"))

	mux := svc.ServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(grpcCred.ClientCred())}
	err := greet_api.RegisterGreetServiceHandlerFromEndpoint(svc.Options().Context, mux, endPoint, opts)
	if err != nil {
		return err
	}
	return svc.Run()
}
