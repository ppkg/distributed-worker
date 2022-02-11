package dto

import "github.com/ppkg/distributed-worker/enum"

// job通知信息
type JobNotify struct {
	// job ID
	Id int64 `json:"id"`
	// job名称
	Name string `json:"name"`
	// meta信息
	Meta map[string]string `json:"meta"`
	// job状态,2：执行完成，3：取消执行，4：系统异常，5：task推送失败，6：运行超时，7：业务处理异常
	Status enum.JobStatus `json:"status"`
	// 结果输出
	Result string `json:"result"`
	// 错误信息
	Message string `json:"message"`
}

type SyncJobRequest struct {
	// job名称
	Name string `json:"name"`
	// meta信息
	Meta map[string]string `json:"meta"`
	// job标签，便于job快速搜索
	Label string `json:"label"`
	// task处理插件集合
	PluginSet []string `json:"plugin_set"`
	// task入参
	TaskInputList []string `json:"-"`
	// task异常操作，0：退出job执行，1：跳过当前task继续执行下一个
	TaskExceptionOperation enum.TaskExceptionOperation `json:"task_exception_operation"`
}

type SyncJobResponse struct {
	// job ID
	Id int64 `json:"id"`
	// meta信息
	Meta map[string]string `json:"meta"`
	// job状态,0:待执行，1：执行中，2：执行完成，3：取消执行，4：系统异常，5：推送失败，6：运行超时，7：业务处理异常
	Status enum.JobStatus `json:"status"`
	// 处理结果
	Result string `json:"result"`
	// 错误信息
	Message string `json:"message"`
}

type AsyncJobRequest struct {
	// job名称
	Name string `json:"name"`
	// meta信息
	Meta map[string]string `json:"meta"`
	// job标签，便于job快速搜索
	Label string `json:"label"`
	// job类型,异步回调通知时使用，根据不同值执行对应业务
	Type string `json:"type"`
	// job完成是否需要通知
	IsNotify bool `json:"is_notify"`
	// task处理插件集合
	PluginSet []string `json:"plugin_set"`
	// task入参
	TaskInputList []string `json:"-"`
	// task异常操作，0：退出job执行，1：跳过当前task继续执行下一个
	TaskExceptionOperation enum.TaskExceptionOperation `json:"task_exception_operation"`
}

type StartNotify struct {
	// job ID
	Id int64 `json:"id"`
	// job名称
	Name string `json:"name"`
	// meta信息
	Meta map[string]string `json:"meta"`
}
