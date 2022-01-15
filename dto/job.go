package dto

// job通知信息
type JobNotify struct {
	// job ID
	Id int64
	// job名称
	Name string
	// job状态,2：执行完成，3：取消执行，4：系统异常，5：task推送失败，6：运行超时，7：业务处理异常
	Status int32
	// 结果输出
	Result string
}
