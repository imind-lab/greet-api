GOPATH := $(shell go env GOPATH)
VERSION := 0.0.1.0

gengo:
	protoc -I. --proto_path ../application/greeter-api/proto --proto_path ../pkg/proto \
 --go_out ../application/greeter-api/proto --go_opt paths=source_relative --go-grpc_out ../application/greeter-api/proto --go-grpc_opt paths=source_relative \
 --grpc-gateway_out ../application/greeter-api/proto --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true greeter-api.proto
	@microctl inject --path=../application/greeter-api/proto/greeter-api.pb.go
	@sed -i '' 's/,omitempty//g' ../application/greeter-api/proto/greeter-api.pb.go

depend:
	go get ../...

build: depend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o greeter ../main.go

test:
	go test -v ../... -cover

docker: gengo
	docker build -f ./Dockerfile -t registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION) ../
	# docker push registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION)


deploy: docker
	@helm upgrade --install reviews ../deploy/helm/greeter-api --set image.tag=$(VERSION)

clean:
	docker rmi registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION)

k8s: docker
	kubectl set image deployment/greeter-api greeter=registry.cn-beijing.aliyuncs.com/imind/greeter-api:$(VERSION)

.PHONY: gengo depend build test docker deploy clean k8s
