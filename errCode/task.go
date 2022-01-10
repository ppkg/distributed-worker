package errCode

const (
	ErrPluginUnsupport = 100201
)

func init() {
	regErrCode(map[int]string{
		ErrPluginUnsupport: "Plugin插件(%s)系统不支持",
	})
}
