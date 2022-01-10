package main

import (
	"context"
	"distributed-worker/core"
	"distributed-worker/proto/job"
	"distributed-worker/proto/task"
	"distributed-worker/service"
	"flag"
	"fmt"
	"time"

	"github.com/ppkg/kit"
	"google.golang.org/grpc"
)

var (
	port        = flag.Int("port", 50051, "TCP port for this node")
	isSubmitJob = flag.Bool("isSubmitJob", false, "is commit job")
)

func main() {
	flag.Parse()

	app := core.NewApp(core.WithAppNameOption("distributed-worker"), core.WithPortOption(*port), core.WithSchedulerUrlOption("127.0.0.1:5001"))
	app.RegisterPlugin(core.NewPlus()).RegisterGrpc(func(appCtx *core.ApplicationContext, server *grpc.Server) {
		task.RegisterTaskServiceServer(server, service.NewTaskService(appCtx))
	})

	if *isSubmitJob {
		submit(app)
	}

	err := app.Run()
	if err != nil {
		fmt.Println("got err:", err)
	}
}

func submit(ctx *core.ApplicationContext) {
	time.Sleep(10 * time.Second)
	fmt.Println("开始创建任务...")
	client := job.NewJobServiceClient(ctx.GetMasterConn())
	req, err := client.SyncSubmit(context.Background())
	if err != nil {
		fmt.Println("请求SyncSubmit出错 ", err)
		return
	}
	data := []map[string]interface{}{
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
		{
			"a": 1,
			"b": 2,
		},
	}
	for _, item := range data {
		req.Send(&job.SubmitRequest{
			Name: "plus-test",
			Type: "plusJob",
			PluginSet: []string{
				"plus",
			},
			Data: kit.JsonEncode(item),
		})
	}

	resp, err := req.CloseAndRecv()
	if err != nil {
		fmt.Println("获取结果异常", err)
		return
	}

	fmt.Println("请求处理成功：", resp.Id, resp.Status, resp.Result)

}
