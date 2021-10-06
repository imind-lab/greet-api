module github.com/imind-lab/greeter-api

go 1.16

replace (
	github.com/imind-lab/greeter => ../../../github.com/imind-lab/greeter
	github.com/imind-lab/micro => ../../../github.com/imind-lab/micro
)

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alibaba/sentinel-golang v1.0.2
	github.com/go-playground/validator/v10 v10.9.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.6.0
	github.com/imind-lab/greeter v0.0.0-00010101000000-000000000000
	github.com/imind-lab/micro v0.0.0-20210930155647-5602387874f9
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/spiffe/go-spiffe/v2 v2.0.0-beta.8
	go.uber.org/zap v1.19.1
	google.golang.org/genproto v0.0.0-20210929214142-896c89f843d2
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)
