package core

import (
	"time"

	"github.com/ppkg/kit"
)

type plusPlugin struct {
}

func (s *plusPlugin) Name() string {
	return "plus"
}

func (s *plusPlugin) Execute(Id int64, jobId int64, input string) (string, error) {
	var params plusParam
	kit.JsonDecode([]byte(input), &params)
	time.Sleep(3 * time.Second)
	return kit.JsonEncode(map[string]interface{}{
		"result": params.A + params.B,
	}), nil
}

func NewPlus() Plugin {
	return &plusPlugin{}
}

type plusParam struct {
	A int
	B int
}
