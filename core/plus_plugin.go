package core

import (
	"fmt"
	"time"

	"github.com/ppkg/kit"
)

type plusPlugin struct {
}

func (s *plusPlugin) Name() string {
	return "plus"
}

func (s *plusPlugin) Handle(ctx *ApplicationContext, Id int64, jobId int64, input string) (string, error) {
	var params plusParam
	_ = kit.JsonDecode([]byte(input), &params)
	time.Sleep(10 * time.Second)
	fmt.Printf("%s->完成任务(%d,%d)\n", s.Name(), Id, jobId)
	return kit.JsonEncode(map[string]interface{}{
		"result": params.A + params.B,
	}), nil
}

func NewPlus() PluginHandler {
	return &plusPlugin{}
}

type plusParam struct {
	A int
	B int
}
