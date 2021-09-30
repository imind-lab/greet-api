/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package service

import (
	greet_api "github.com/imind-lab/greet-api/server/proto/greet-api"
	"github.com/imind-lab/greet/server/proto/greet"
)

func GreetMap(pos []*greet.Greet, fn func(*greet.Greet) *greet_api.Greet) []*greet_api.Greet {
	var dtos []*greet_api.Greet
	for _, po := range pos {
		dtos = append(dtos, fn(po))
	}
	return dtos
}

func GreetGw2Srv(po *greet_api.Greet) *greet.Greet {
	if po == nil {
		return nil
	}

	dto := &greet.Greet{}
	dto.Id = po.Id
	dto.Name = po.Name
	dto.ViewNum = po.ViewNum
	dto.Status = po.Status
	dto.CreateTime = po.CreateTime
	dto.UpdateDatetime = po.UpdateDatetime
	dto.CreateDatetime = po.CreateDatetime

	return dto
}

func GreetSrv2Gw(dto *greet.Greet) *greet_api.Greet {
	if dto == nil {
		return nil
	}

	po := &greet_api.Greet{}
	po.Id = dto.Id
	po.Name = dto.Name
	po.ViewNum = dto.ViewNum
	po.Status = dto.Status
	po.CreateTime = dto.CreateTime
	po.UpdateDatetime = dto.UpdateDatetime
	po.CreateDatetime = dto.CreateDatetime

	return po
}

func GreetListSrv2Gw(dto *greet.GreetList) *greet_api.GreetList {
	if dto == nil {
		return nil
	}

	po := &greet_api.GreetList{}
	po.Total = dto.Total
	po.TotalPage = dto.TotalPage
	po.CurPage = dto.CurPage
	po.Datalist = GreetMap(dto.Datalist, GreetSrv2Gw)

	return po
}

func ErrorSrv2Gw(err *greet.Error) *greet_api.Error {
	if err == nil {
		return nil
	}

	po := &greet_api.Error{}
	po.Message = err.Message
	po.Code = err.Code

	return po
}
