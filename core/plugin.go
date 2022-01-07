package core

type Plugin interface {
	Name() string
	Execute(Id, jobId int64, input string) (string, error)
}
