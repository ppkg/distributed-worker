package service

import (
	"context"
	"distributed-worker/core"
	"distributed-worker/enum"
	"distributed-worker/errCode"
	"distributed-worker/proto/task"
	"fmt"

	"github.com/ppkg/glog"
	"github.com/ppkg/kit"
)

type taskService struct {
	appCtx *core.ApplicationContext
}

// 同步提交task
func (s *taskService) SyncSubmit(ctx context.Context, req *task.SubmitRequest) (*task.SyncSubmitResponse, error) {
	handler := s.appCtx.GetPluginHandler(req.Plugin)
	if handler == nil {
		glog.Errorf("taskService/SyncSubmit 当前服务不支持插件%s,请求参数:%s", req.Plugin, kit.JsonEncode(req))
		return nil, errCode.ToGrpcErr(errCode.ErrPluginUnsupport, req.Plugin)
	}
	resp := &task.SyncSubmitResponse{
		Id:     req.Id,
		JobId:  req.JobId,
		Status: enum.FinishTaskStatus,
	}
	result, err := handler.Handle(s.appCtx,req.Id, req.JobId, req.Data)
	if err != nil {
		glog.Errorf("taskService/SyncSubmit 运行插件%s异常,请求参数:%s,err:%+v", req.Plugin, kit.JsonEncode(req), err)
		resp.Status = enum.ExceptionTaskStatus
		resp.Result = fmt.Sprintf("运行插件%s异常,err:%+v", req.Plugin, err)
		return resp, nil
	}
	resp.Result = result
	return resp, nil
}

func NewTaskService(appCtx *core.ApplicationContext) task.TaskServiceServer {
	return &taskService{
		appCtx: appCtx,
	}
}
