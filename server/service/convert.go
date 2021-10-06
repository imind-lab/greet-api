/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package service

import (
	greeter_api "github.com/imind-lab/greeter-api/server/proto/greeter-api"
	"github.com/imind-lab/greeter/server/proto/greeter"
)

func GreeterMap(pos []*greeter.Greeter, fn func(*greeter.Greeter) *greeter_api.Greeter) []*greeter_api.Greeter {
	var dtos []*greeter_api.Greeter
	for _, po := range pos {
		dtos = append(dtos, fn(po))
	}
	return dtos
}

func GreeterGw2Srv(po *greeter_api.Greeter) *greeter.Greeter {
	if po == nil {
		return nil
	}

	dto := &greeter.Greeter{}
	dto.Id = po.Id
	dto.Name = po.Name
	dto.ViewNum = po.ViewNum
	dto.Status = po.Status
	dto.CreateTime = po.CreateTime
	dto.UpdateDatetime = po.UpdateDatetime
	dto.CreateDatetime = po.CreateDatetime

	return dto
}

func GreeterSrv2Gw(dto *greeter.Greeter) *greeter_api.Greeter {
	if dto == nil {
		return nil
	}

	po := &greeter_api.Greeter{}
	po.Id = dto.Id
	po.Name = dto.Name
	po.ViewNum = dto.ViewNum
	po.Status = dto.Status
	po.CreateTime = dto.CreateTime
	po.UpdateDatetime = dto.UpdateDatetime
	po.CreateDatetime = dto.CreateDatetime

	return po
}

func GreeterListSrv2Gw(dto *greeter.GreeterList) *greeter_api.GreeterList {
	if dto == nil {
		return nil
	}

	po := &greeter_api.GreeterList{}
	po.Total = dto.Total
	po.TotalPage = dto.TotalPage
	po.CurPage = dto.CurPage
	po.Datalist = GreeterMap(dto.Datalist, GreeterSrv2Gw)

	return po
}

func ErrorSrv2Gw(err *greeter.Error) *greeter_api.Error {
	if err == nil {
		return nil
	}

	po := &greeter_api.Error{}
	po.Message = err.Message
	po.Code = err.Code

	return po
}
