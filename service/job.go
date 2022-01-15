package service

import (
	"context"
	"distributed-worker/core"
	"distributed-worker/dto"
	"distributed-worker/errCode"
	"distributed-worker/proto/job"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ppkg/glog"
	"github.com/ppkg/kit"
)

type jobService struct {
	appCtx *core.ApplicationContext
}

// 异步提交job
func (s *jobService) AsyncSubmit(_ job.JobService_AsyncSubmitServer) error {
	panic("not implemented")
}

// 异步通知
func (s *jobService) AsyncNotify(ctx context.Context, req *job.AsyncNotifyRequest) (*empty.Empty, error) {
	handler := s.appCtx.GetJobNotifyHandler(req.Type)
	if handler == nil {
		glog.Errorf("taskService/AsyncNotify 当前服务不支持该通知类型(%s),请求参数:%s", req.Type, kit.JsonEncode(req))
		return nil, errCode.ToGrpcErr(errCode.ErrJobNotifyUnsupport, req.Type)
	}
	jobNotify := dto.JobNotify{
		Id:     req.Id,
		Name:   req.Name,
		Status: req.Status,
		Result: req.Result,
	}
	err := handler.Handle(jobNotify)
	if err != nil {
		glog.Errorf("taskService/AsyncNotify 运行通知处理器(%s)异常,请求参数:%s,err:%+v", req.Type, kit.JsonEncode(req), err)
		return nil, err
	}
	return &empty.Empty{}, nil

}

// 同步提交job(当调度器挂掉会导致job处理中断，谨用)
func (s *jobService) SyncSubmit(_ job.JobService_SyncSubmitServer) error {
	panic("not implemented")
}

func NewJobService() job.JobServiceServer {
	return &jobService{}
}
