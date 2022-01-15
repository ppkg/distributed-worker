package core

import (
	"fmt"
	"time"

	"github.com/ppkg/kit"
)

type multiPlugin struct {
}

func (s *multiPlugin) Name() string {
	return "multi"
}

func (s *multiPlugin) Handle(Id int64, jobId int64, input string) (string, error) {
	var params multiParam
	_=kit.JsonDecode([]byte(input), &params)
	time.Sleep(5 * time.Second)
	fmt.Printf("%s->完成任务(%d,%d)\n", s.Name(), Id, jobId)
	return kit.JsonEncode(map[string]interface{}{
		"result": params.Result * 10,
	}), nil
}

func NewMulti() PluginHandler {
	return &multiPlugin{}
}

type multiParam struct {
	Result int
}
