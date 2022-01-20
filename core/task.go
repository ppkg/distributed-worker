package core

import (
	"context"
	"fmt"

	"github.com/ppkg/distributed-worker/enum"
	"github.com/ppkg/distributed-worker/errCode"
	"github.com/ppkg/distributed-worker/proto/task"

	"github.com/ppkg/glog"
	"github.com/ppkg/kit"
)

type taskService struct {
	appCtx *ApplicationContext
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
		Status: int32(enum.FinishTaskStatus),
	}
	result, err := handler.Handle(req.Id, req.JobId, req.Data)
	if err != nil {
		glog.Errorf("taskService/SyncSubmit 运行插件%s异常,请求参数:%s,err:%+v", req.Plugin, kit.JsonEncode(req), err)
		resp.Status = int32(enum.ExceptionTaskStatus)
		resp.Message = fmt.Sprintf("运行插件%s异常,err:%+v", req.Plugin, err)
		return resp, nil
	}
	resp.Result = result
	return resp, nil
}

func NewTaskService(appCtx *ApplicationContext) task.TaskServiceServer {
	return &taskService{
		appCtx: appCtx,
	}
}
