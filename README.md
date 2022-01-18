# distributed-worker 分布式调度接入指南

distributed-worker是一个由Go开发高性能分布式调度项目，distributed-worker在分布式环境上调度job功能上，增强系统容错性，支持随时增减服务实例节点，能够加快数据运算，提升系统之间交互效率



## 安装

参考demo：

[distributed-worker-demo](https://github.com/ppkg/distributed-worker-demo)

安装依赖包:

```
go get github.com/ppkg/distributed-worker
```



## 初始化代码

```go
	app = core.NewApp(core.WithNacosSchedulerServiceNameOption("distributed-scheduler"), core.WithNacosAddrOption("mse-e52dbdd6-p.nacos-ans.mse.aliyuncs.com:8848"), core.WithNacosNamespaceOption("27fdefc2-ae39-41fd-bac4-9256acbf97bc"), core.WithNacosServiceGroupOption("my-service"))

	// 注册task处理插件
	app.RegisterPlugin(func(ctx *core.ApplicationContext) core.PluginHandler {
		return handler.NewPlus()
	}).RegisterPlugin(func(ctx *core.ApplicationContext) core.PluginHandler {
		return handler.NewMulti()
	})

	// 注册回调通知
	app.RegisterJobNotify(func(ctx *core.ApplicationContext) core.JobNotifyHandler {
		return handler.NewDemoNotify()
	})

	err := app.Run()
	if err != nil {
		fmt.Println("start app got err:", err)
	}
```



## 可配置参数

| 可选参数option                      | 环境变量                     | 是否必填 | 参数说明                                                     |
| ----------------------------------- | ---------------------------- | -------- | ------------------------------------------------------------ |
| WithNacosAddrOption                 | NACOS_ADDRS                  | 是       | nacos服务地址(包含端口号)，环境变量支持多个nacos服务，需要以“,”分隔 |
| WithNacosNamespaceOption            | NACOS_NAMESPACE              | 否       | nacos命名空间，为空则为nacos默认命名空间                     |
| WithNacosServiceGroupOption         | NACOS_SERVICE_GROUP          | 否       | nacos分组，为空则为nacos默认分组                             |
| WithNacosClusterNameOption          | NACOS_CLUSTER_NAME           | 否       | 指定nacos集群，为空则为nacos默认集群                         |
| WithNacosSchedulerServiceNameOption | NACOS_SCHEDULER_SERVICE_NAME | 否       | 指定调度器服务名，默认值：distributed-scheduler              |
| WithAppNameOption                   | APP_NAME                     | 否       | 指定应用名称，默认为空                                       |
| WithPortOption                      | APP_PORT                     | 否       | 指定服务端口号，默认值：8080                                 |
| WithEndpointOption                  | ENDPOINT                     | 否       | 指定对外服务端点地址，默认当前ip地址                         |

