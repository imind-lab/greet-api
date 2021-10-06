/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	greeter_api "github.com/imind-lab/greeter-api/server/proto/greeter-api"
	greeterClient "github.com/imind-lab/greeter/client"
	"github.com/imind-lab/greeter/server/proto/greeter"
	"github.com/imind-lab/micro/sentinelx"
)

type GreeterService struct {
	greeter_api.UnimplementedGreeterServiceServer

	validate *validator.Validate

	ds *sentinelx.Sentinel
}

//NewGreeterService 创建用户服务实例
func NewGreeterService(logger *zap.Logger) *GreeterService {
	ds, _ := sentinelx.NewSentinel(logger)
	svc := &GreeterService{
		ds:       ds,
		validate: validator.New(),
	}
	return svc
}

// CreateGreeter 创建Greeter
func (svc *GreeterService) CreateGreeter(ctx context.Context, req *greeter_api.CreateGreeterRequest) (*greeter_api.CreateGreeterResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "CreateGreeter"))
	logger.Debug("Receive CreateGreeter request")

	rsp := &greeter_api.CreateGreeterResponse{}

	m := req.Dto
	fmt.Println("Dto", m)
	err := svc.validate.Struct(req)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
		}

		for _, err := range err.(validator.ValidationErrors) {

			fmt.Println(err.Namespace())
			fmt.Println(err.Field())
			fmt.Println(err.StructNamespace())
			fmt.Println(err.StructField())
			fmt.Println(err.Tag())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println()
		}

	}
	err = svc.validate.Var(m, "required")
	fmt.Println("validate", err)
	if m == nil {
		logger.Error("Greeter不能为空", zap.Any("params", m))

		err := &greeter_api.Error{}
		err.Message = "Greeter不能为空"
		rsp.Error = err
		return rsp, nil
	}

	err = svc.validate.Var(m.Name, "required,email")
	fmt.Println("validate", m.Name, err)
	if len(m.Name) == 0 {
		logger.Error("Name不能为空", zap.Any("name", m.Name))

		err := &greeter_api.Error{}
		err.Message = "Name不能为空"
		rsp.Error = err
		return rsp, nil
	}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "CreateGreeter Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("服务器繁忙，请稍候再试", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}
	greeterSrv := GreeterGw2Srv(req.Dto)
	resp, err := greeterCli.CreateGreeter(ctx, &greeter.CreateGreeterRequest{
		Dto: greeterSrv,
	})
	if err != nil {
		logger.Error("greeterCli.CreateGreeter error", zap.Any("greeter", m), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "创建Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)
	return rsp, nil
}

// GetGreeterById 根据Id获取Greeter
func (svc *GreeterService) GetGreeterById(ctx context.Context, req *greeter_api.GetGreeterByIdRequest) (*greeter_api.GetGreeterByIdResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "GetGreeterById"))
	logger.Debug("Receive GetGreeterById request")

	rsp := &greeter_api.GetGreeterByIdResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "GetGreeterById Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)
	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("greeterClient.New error", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	sentinelEntry, blockError := sentinel.Entry("test1")
	if blockError != nil {
		logger.Error("触发熔断降级", zap.Any("TriggeredRule", blockError.TriggeredRule()), zap.Any("TriggeredValue", blockError.TriggeredValue()))

		err := &greeter_api.Error{}
		err.Message = "获取Greeter失败"
		rsp.Error = err
		return rsp, nil
	}
	defer sentinelEntry.Exit()

	resp, err := greeterCli.GetGreeterById(ctx, &greeter.GetGreeterByIdRequest{
		Id: req.Id,
	})
	ctxzap.Debug(ctx, "greeterCli.GetGreeterById", zap.Any("resp", resp), zap.Error(err))
	if err != nil {
		logger.Error("greeterCli.GetGreeterById error", zap.Any("resp", resp), zap.Error(err))

		sentinelEntry.SetError(err)

		err := &greeter_api.Error{}
		err.Message = "获取Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Dto = GreeterSrv2Gw(resp.Dto)
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}

func (svc *GreeterService) GetGreeterList(ctx context.Context, req *greeter_api.GetGreeterListRequest) (*greeter_api.GetGreeterListResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "GetGreeterList"))
	logger.Debug("Receive GetGreeterList request")

	rsp := &greeter_api.GetGreeterListResponse{}

	err := svc.validate.Struct(req)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
		}

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.Namespace())
			fmt.Println(err.Field())
			fmt.Println(err.StructNamespace())
			fmt.Println(err.StructField())
			fmt.Println(err.Tag())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println()
		}

	}

	rateEntry, rateError := sentinel.Entry("abcd", sentinel.WithTrafficType(base.Inbound))
	if rateError != nil {
		ctxzap.Debug(ctx, "GetGreeterList限流了")
		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}
	defer rateEntry.Exit()
	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "GetGreeterList Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("greeterClient.New error", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greeterCli.GetGreeterList(ctx, &greeter.GetGreeterListRequest{
		Status:   req.Status,
		Lastid:   req.Lastid,
		Pagesize: req.Pagesize,
		Page:     req.Page,
	})
	if err != nil {
		logger.Error("greeterCli.GetGreeterList error", zap.Any("resp", resp), zap.Error(err))
		err := &greeter_api.Error{}
		err.Message = "获取GreeterList失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Data = GreeterListSrv2Gw(resp.Data)
	rsp.Error = ErrorSrv2Gw(resp.Error)
	return rsp, nil
}

func (svc *GreeterService) UpdateGreeterStatus(ctx context.Context, req *greeter_api.UpdateGreeterStatusRequest) (*greeter_api.UpdateGreeterStatusResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "UpdateGreeterStatus"))
	logger.Debug("Receive UpdateGreeterStatus request")

	rsp := &greeter_api.UpdateGreeterStatusResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "UpdateGreeterStatus Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("greeterClient.New error", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greeterCli.UpdateGreeterStatus(ctx, &greeter.UpdateGreeterStatusRequest{
		Id:     req.Id,
		Status: req.Status,
	})
	if err != nil {
		logger.Error("greeterCli.UpdateGreeterStatus error", zap.Any("resp", resp), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "更新Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}

func (svc *GreeterService) UpdateGreeterCount(ctx context.Context, req *greeter_api.UpdateGreeterCountRequest) (*greeter_api.UpdateGreeterCountResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "UpdateGreeterCount"))
	logger.Debug("Receive UpdateGreeterCount request")

	rsp := &greeter_api.UpdateGreeterCountResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "UpdateGreeterCount Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("greeterClient.New error", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greeterCli.UpdateGreeterCount(ctx, &greeter.UpdateGreeterCountRequest{
		Id:     req.Id,
		Num:    req.Num,
		Column: req.Column,
	})
	if err != nil {
		logger.Error("greeterCli.UpdateGreeterCount error", zap.Any("resp", resp), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "更新Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}

func (svc *GreeterService) DeleteGreeterById(ctx context.Context, req *greeter_api.DeleteGreeterByIdRequest) (*greeter_api.DeleteGreeterByIdResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "DeleteGreeterById"))
	logger.Debug("Receive DeleteGreeterById request")

	rsp := &greeter_api.DeleteGreeterByIdResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "DeleteGreeterById Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("greeterClient.New error", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greeterCli.DeleteGreeterById(ctx, &greeter.DeleteGreeterByIdRequest{
		Id: req.Id,
	})
	if err != nil {
		logger.Error("greeterCli.DeleteGreeterById error", zap.Any("resp", resp), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "删除Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}
func (svc *GreeterService) GetGreeterListByIds(ctx context.Context, req *greeter_api.GetGreeterListByIdsRequest) (*greeter_api.GetGreeterListByIdsResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "GetGreeterListByIds"))
	logger.Debug("Receive GetGreeterListByIds request")

	rsp := &greeter_api.GetGreeterListByIdsResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "GetGreeterListByIds Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greeterCli, err := greeterClient.New(ctx)
	if err != nil {
		logger.Error("greeterClient.New error", zap.Any("greeterCli", greeterCli), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	data := make([]*greeter_api.Greeter, len(req.Ids))

	streamClient, err := greeterCli.GetGreeterListByStream(ctx)
	if err != nil {
		logger.Error("greeterCli.GetGreeterListByStream error", zap.Any("streamClient", streamClient), zap.Error(err))

		err := &greeter_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			resp, err := streamClient.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				ctxzap.Error(ctx, "GetGreeterListByStream Recv error", zap.Error(err))
				return
			}
			fmt.Println("Recv", resp.Index, resp.Result)
			data[resp.Index] = GreeterSrv2Gw(resp.Result)
		}
	}()

	for key, val := range req.Ids {
		_ = streamClient.Send(&greeter.GetGreeterListByStreamRequest{
			Index: int32(key),
			Id:    val,
		})
	}
	streamClient.CloseSend()
	wg.Wait()

	for _, m := range data {
		if m != nil {
			rsp.Data = append(rsp.Data, m)
		}
	}

	rsp.Success = true

	return rsp, nil
}

func (svc *GreeterService) Close() {
	if svc.ds != nil {
		svc.ds.Close()
	}
}
