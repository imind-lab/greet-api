GOPATH := $(shell go env GOPATH)
VERSION := 0.0.1.0

gengo:
	protoc -I. --proto_path ../server/proto \
 --go_out ../server/proto --go_opt paths=source_relative --go-grpc_out ../server/proto --go-grpc_opt paths=source_relative \
 --grpc-gateway_out ../server/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true greet-api/greet-api.proto
	protoc-go-inject-tag -input=../server/proto/greet-api/greet-api.pb.go

genphp:
	protoc -I. --proto_path ../server/proto \
 --php_out ../server/proto/greet-api --grpc_out ../server/proto/greet-api --plugin=protoc-gen-grpc=${GOPATH}/bin/grpc_php_plugin greet-api/greet-api.proto

depend: gengo
	go get ../...

build: depend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o greet ../main.go

test:
	go test -v ../... -cover

docker: build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o grpc-health-probe ../pkg/grpc-health-probe/main.go
	docker build -f ./Dockerfile -t registry.cn-beijing.aliyuncs.com/imind/greet-api:$(VERSION) ../
	#docker push registry.cn-beijing.aliyuncs.com/imind/greet-api:$(VERSION)
	rm -rf grpc-health-probe
	rm -rf greet

deploy: docker
	kubectl apply -f ns-rbac.yaml
	kubectl apply -f greet-api.yaml

clean:
	docker rmi registry.cn-beijing.aliyuncs.com/imind/greet-api:$(VERSION)

k8s: docker
	kubectl set image deployment/greet-api greet=registry.cn-beijing.aliyuncs.com/imind/greet-api:$(VERSION) -n micro

.PHONY: gengo genphp depend build test docker deploy clean k8s
