package dto

import "github.com/ppkg/distributed-worker/enum"

// job通知信息
type JobNotify struct {
	// job ID
	Id int64
	// job名称
	Name string
	// meta信息
	Meta map[string]string
	// job状态,2：执行完成，3：取消执行，4：系统异常，5：task推送失败，6：运行超时，7：业务处理异常
	Status enum.JobStatus
	// 结果输出
	Result string
	// 错误信息
	Message string
}

type SyncJobRequest struct {
	// job名称
	Name string
	// job标签，便于job快速搜索
	Label string
	// task处理插件集合
	PluginSet []string
	// task入参
	TaskInputList []string `json:"-"`
	// task异常操作，0：退出job执行，1：跳过当前task继续执行下一个
	TaskExceptionOperation enum.TaskExceptionOperation
}

type SyncJobResponse struct {
	// job ID
	Id int64
	// job状态,0:待执行，1：执行中，2：执行完成，3：取消执行，4：系统异常，5：推送失败，6：运行超时，7：业务处理异常
	Status enum.JobStatus
	// 处理结果
	Result string
	// 错误信息
	Message string
}

type AsyncJobRequest struct {
	// job名称
	Name string
	// job标签，便于job快速搜索
	Label string
	// job类型,异步回调通知时使用，根据不同值执行对应业务
	Type string
	// job完成是否需要通知
	IsNotify bool
	// task处理插件集合
	PluginSet []string
	// task入参
	TaskInputList []string `json:"-"`
	// task异常操作，0：退出job执行，1：跳过当前task继续执行下一个
	TaskExceptionOperation enum.TaskExceptionOperation
}

type StartNotify struct {
	// job ID
	Id int64
	// job名称
	Name string
	// meta信息
	Meta map[string]string
}
