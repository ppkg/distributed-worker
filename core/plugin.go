package core

type PluginHandler interface {
	Name() string
	Handle(Id, jobId int64, input string) (string, error)
}
