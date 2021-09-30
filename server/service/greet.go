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

	greet_api "github.com/imind-lab/greet-api/server/proto/greet-api"
	greetClient "github.com/imind-lab/greet/client"
	"github.com/imind-lab/greet/server/proto/greet"
	"github.com/imind-lab/micro/sentinelx"
)

type GreetService struct {
	greet_api.UnimplementedGreetServiceServer

	validate *validator.Validate

	ds *sentinelx.Sentinel
}

//NewGreetService 创建用户服务实例
func NewGreetService(logger *zap.Logger) *GreetService {
	ds, _ := sentinelx.NewSentinel(logger)
	svc := &GreetService{
		ds:       ds,
		validate: validator.New(),
	}
	return svc
}

// CreateGreet 创建Greet
func (svc *GreetService) CreateGreet(ctx context.Context, req *greet_api.CreateGreetRequest) (*greet_api.CreateGreetResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "CreateGreet"))
	logger.Debug("Receive CreateGreet request")

	rsp := &greet_api.CreateGreetResponse{}

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
		logger.Error("Greet不能为空", zap.Any("params", m))

		err := &greet_api.Error{}
		err.Message = "Greet不能为空"
		rsp.Error = err
		return rsp, nil
	}

	err = svc.validate.Var(m.Name, "required,email")
	fmt.Println("validate", m.Name, err)
	if len(m.Name) == 0 {
		logger.Error("Name不能为空", zap.Any("name", m.Name))

		err := &greet_api.Error{}
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
	ctxzap.Debug(ctx, "CreateGreet Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("服务器繁忙，请稍候再试", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}
	greetSrv := GreetGw2Srv(req.Dto)
	resp, err := greetCli.CreateGreet(ctx, &greet.CreateGreetRequest{
		Dto: greetSrv,
	})
	if err != nil {
		logger.Error("greetCli.CreateGreet error", zap.Any("greet", m), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "创建Greet失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)
	return rsp, nil
}

// GetGreetById 根据Id获取Greet
func (svc *GreetService) GetGreetById(ctx context.Context, req *greet_api.GetGreetByIdRequest) (*greet_api.GetGreetByIdResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "GetGreetById"))
	logger.Debug("Receive GetGreetById request")

	rsp := &greet_api.GetGreetByIdResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "GetGreetById Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)
	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("greetClient.New error", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	sentinelEntry, blockError := sentinel.Entry("test1")
	if blockError != nil {
		logger.Error("触发熔断降级", zap.Any("TriggeredRule", blockError.TriggeredRule()), zap.Any("TriggeredValue", blockError.TriggeredValue()))

		err := &greet_api.Error{}
		err.Message = "获取Greet失败"
		rsp.Error = err
		return rsp, nil
	}
	defer sentinelEntry.Exit()

	resp, err := greetCli.GetGreetById(ctx, &greet.GetGreetByIdRequest{
		Id: req.Id,
	})
	ctxzap.Debug(ctx, "greetCli.GetGreetById", zap.Any("resp", resp), zap.Error(err))
	if err != nil {
		logger.Error("greetCli.GetGreetById error", zap.Any("resp", resp), zap.Error(err))

		sentinelEntry.SetError(err)

		err := &greet_api.Error{}
		err.Message = "获取Greet失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Dto = GreetSrv2Gw(resp.Dto)
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}

func (svc *GreetService) GetGreetList(ctx context.Context, req *greet_api.GetGreetListRequest) (*greet_api.GetGreetListResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "GetGreetList"))
	logger.Debug("Receive GetGreetList request")

	rsp := &greet_api.GetGreetListResponse{}

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
		ctxzap.Debug(ctx, "GetGreetList限流了")
		err := &greet_api.Error{}
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
	ctxzap.Debug(ctx, "GetGreetList Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("greetClient.New error", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greetCli.GetGreetList(ctx, &greet.GetGreetListRequest{
		Status:   req.Status,
		Lastid:   req.Lastid,
		Pagesize: req.Pagesize,
		Page:     req.Page,
	})
	if err != nil {
		logger.Error("greetCli.GetGreetList error", zap.Any("resp", resp), zap.Error(err))
		err := &greet_api.Error{}
		err.Message = "获取GreetList失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Data = GreetListSrv2Gw(resp.Data)
	rsp.Error = ErrorSrv2Gw(resp.Error)
	return rsp, nil
}

func (svc *GreetService) UpdateGreetStatus(ctx context.Context, req *greet_api.UpdateGreetStatusRequest) (*greet_api.UpdateGreetStatusResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "UpdateGreetStatus"))
	logger.Debug("Receive UpdateGreetStatus request")

	rsp := &greet_api.UpdateGreetStatusResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "UpdateGreetStatus Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("greetClient.New error", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greetCli.UpdateGreetStatus(ctx, &greet.UpdateGreetStatusRequest{
		Id:     req.Id,
		Status: req.Status,
	})
	if err != nil {
		logger.Error("greetCli.UpdateGreetStatus error", zap.Any("resp", resp), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "更新Greet失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}

func (svc *GreetService) UpdateGreetCount(ctx context.Context, req *greet_api.UpdateGreetCountRequest) (*greet_api.UpdateGreetCountResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "UpdateGreetCount"))
	logger.Debug("Receive UpdateGreetCount request")

	rsp := &greet_api.UpdateGreetCountResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "UpdateGreetCount Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("greetClient.New error", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greetCli.UpdateGreetCount(ctx, &greet.UpdateGreetCountRequest{
		Id:     req.Id,
		Num:    req.Num,
		Column: req.Column,
	})
	if err != nil {
		logger.Error("greetCli.UpdateGreetCount error", zap.Any("resp", resp), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "更新Greet失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}

func (svc *GreetService) DeleteGreetById(ctx context.Context, req *greet_api.DeleteGreetByIdRequest) (*greet_api.DeleteGreetByIdResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "DeleteGreetById"))
	logger.Debug("Receive DeleteGreetById request")

	rsp := &greet_api.DeleteGreetByIdResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "DeleteGreetById Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("greetClient.New error", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	resp, err := greetCli.DeleteGreetById(ctx, &greet.DeleteGreetByIdRequest{
		Id: req.Id,
	})
	if err != nil {
		logger.Error("greetCli.DeleteGreetById error", zap.Any("resp", resp), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "删除Greet失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = resp.Success
	rsp.Error = ErrorSrv2Gw(resp.Error)

	return rsp, nil
}
func (svc *GreetService) GetGreetListByIds(ctx context.Context, req *greet_api.GetGreetListByIdsRequest) (*greet_api.GetGreetListByIdsResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "GetGreetListByIds"))
	logger.Debug("Receive GetGreetListByIds request")

	rsp := &greet_api.GetGreetListByIdsResponse{}

	uid := 0
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok {
		uids := meta.Get("uid")
		if len(uids) > 0 {
			uid, _ = strconv.Atoi(uids[0])
		}
	}
	ctxzap.Debug(ctx, "GetGreetListByIds Metadata", zap.Any("meta", meta), zap.Int("uid", uid), zap.Bool("ok", ok))

	ctx = metadata.NewOutgoingContext(ctx, meta)

	greetCli, err := greetClient.New(ctx)
	if err != nil {
		logger.Error("greetClient.New error", zap.Any("greetCli", greetCli), zap.Error(err))

		err := &greet_api.Error{}
		err.Message = "服务器繁忙，请稍候再试"
		rsp.Error = err
		return rsp, nil
	}

	data := make([]*greet_api.Greet, len(req.Ids))

	streamClient, err := greetCli.GetGreetListByStream(ctx)
	if err != nil {
		logger.Error("greetCli.GetGreetListByStream error", zap.Any("streamClient", streamClient), zap.Error(err))

		err := &greet_api.Error{}
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
				ctxzap.Error(ctx, "GetGreetListByStream Recv error", zap.Error(err))
				return
			}
			fmt.Println("Recv", resp.Index, resp.Result)
			data[resp.Index] = GreetSrv2Gw(resp.Result)
		}
	}()

	for key, val := range req.Ids {
		_ = streamClient.Send(&greet.GetGreetListByStreamRequest{
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

func (svc *GreetService) Close() {
	if svc.ds != nil {
		svc.ds.Close()
	}
}
