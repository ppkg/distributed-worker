package core

import (
	"context"

	"github.com/ppkg/distributed-worker/dto"
	"github.com/ppkg/distributed-worker/enum"
	"github.com/ppkg/distributed-worker/errCode"
	"github.com/ppkg/distributed-worker/proto/job"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/ppkg/glog"
	"github.com/ppkg/kit"
)

type JobNotifyHandler interface {
	// 返回handler名称，与job中type字段对应
	Name() string
	// job回调通知
	Handle(data dto.JobNotify) error
	// job开始执行通知
	PostStart(data dto.StartNotify) error
}

type jobService struct {
	appCtx *ApplicationContext
}

// 异步提交job
func (s *jobService) AsyncSubmit(_ job.JobService_AsyncSubmitServer) error {
	panic("not implemented")
}

// 异步结果通知
func (s *jobService) AsyncNotify(ctx context.Context, req *job.AsyncNotifyRequest) (*empty.Empty, error) {
	handler := s.appCtx.GetJobNotifyHandler(req.Type)
	if handler == nil {
		glog.Errorf("taskService/AsyncNotify 当前服务不支持该通知类型(%s),请求参数:%s", req.Type, kit.JsonEncode(req))
		return nil, errCode.ToGrpcErr(errCode.ErrJobNotifyUnsupport, req.Type)
	}
	jobNotify := dto.JobNotify{
		Id:      req.Id,
		Name:    req.Name,
		Meta:    req.Meta,
		Status:  enum.JobStatus(req.Status),
		Result:  req.Result,
		Message: req.Mesage,
	}
	err := handler.Handle(jobNotify)
	if err != nil {
		glog.Errorf("taskService/AsyncNotify 运行通知处理器(%s)异常,请求参数:%s,err:%+v", req.Type, kit.JsonEncode(req), err)
		return nil, err
	}
	return &empty.Empty{}, nil

}

// job开始执行通知
func (s *jobService) AsyncPostStart(ctx context.Context, req *job.AsyncPostStartRequest) (*empty.Empty, error) {
	handler := s.appCtx.GetJobNotifyHandler(req.Type)
	if handler == nil {
		glog.Errorf("taskService/AsyncPostStart 当前服务不支持该通知类型(%s),请求参数:%s", req.Type, kit.JsonEncode(req))
		return nil, errCode.ToGrpcErr(errCode.ErrJobNotifyUnsupport, req.Type)
	}
	startNotify := dto.StartNotify{
		Id:   req.Id,
		Name: req.Name,
		Meta: req.Meta,
	}
	err := handler.PostStart(startNotify)
	if err != nil {
		glog.Errorf("taskService/AsyncPostStart 运行通知处理器(%s)异常,请求参数:%s,err:%+v", req.Type, kit.JsonEncode(req), err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

// 同步提交job(当调度器挂掉会导致job处理中断，谨用)
func (s *jobService) SyncSubmit(_ job.JobService_SyncSubmitServer) error {
	panic("not implemented")
}

func NewJobService(appCtx *ApplicationContext) job.JobServiceServer {
	return &jobService{
		appCtx: appCtx,
	}
}
