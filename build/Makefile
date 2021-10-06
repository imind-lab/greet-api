GOPATH := $(shell go env GOPATH)
VERSION := 0.0.1.0

gengo:
	protoc -I. --proto_path ../server/proto \
 --go_out ../server/proto --go_opt paths=source_relative --go-grpc_out ../server/proto --go-grpc_opt paths=source_relative \
 --grpc-gateway_out ../server/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true greeter-api/greeter-api.proto
	protoc-go-inject-tag -input=../server/proto/greeter-api/greeter-api.pb.go

genphp:
	protoc -I. --proto_path ../server/proto \
 --php_out ../server/proto/greeter-api --grpc_out ../server/proto/greeter-api --plugin=protoc-gen-grpc=${GOPATH}/bin/grpc_php_plugin greeter-api/greeter-api.proto

depend: gengo
	go get ../...

build: depend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o greeter ../main.go

test:
	go test -v ../... -cover

docker: build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o grpc-health-probe ../pkg/grpc-health-probe/main.go
	docker build -f ./Dockerfile -t registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION) ../
	#docker push registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION)
	rm -rf grpc-health-probe
	rm -rf greeter

deploy: docker
	kubectl apply -f ns-rbac.yaml
	kubectl apply -f greeter-api.yaml

clean:
	docker rmi registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION)

k8s: docker
	kubectl set image deployment/greeter-api greeter=registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION) -n micro

.PHONY: gengo genphp depend build test docker deploy clean k8s
