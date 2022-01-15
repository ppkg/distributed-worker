package core

type PluginHandler interface {
	Name() string
	Handle(ctx *ApplicationContext, Id, jobId int64, input string) (string, error)
}
