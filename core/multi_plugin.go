package core

import (
	"time"

	"github.com/ppkg/kit"
)

type multiPlugin struct {
}

func (s *multiPlugin) Name() string {
	return "multi"
}

func (s *multiPlugin) Execute(Id int64, jobId int64, input string) (string, error) {
	var params multiParam
	kit.JsonDecode([]byte(input), &params)
	time.Sleep(2 * time.Second)
	return kit.JsonEncode(map[string]interface{}{
		"result": params.Result * 10,
	}), nil
}

func NewMulti() Plugin {
	return &multiPlugin{}
}

type multiParam struct {
	Result int
}
