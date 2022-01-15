package core

import "distributed-worker/dto"

type JobNotifyHandler interface {
	// 返回handler名称，与job中type字段对应
	Name() string
	// 通知业务处理
	Handle(data dto.JobNotify) error
}
