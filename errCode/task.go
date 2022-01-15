package errCode

const (
	ErrPluginUnsupport    = 100201
	ErrJobNotifyUnsupport = 100301
)

func init() {
	regErrCode(map[int]string{
		ErrPluginUnsupport:    "Plugin插件(%s)系统不支持",
		ErrJobNotifyUnsupport: "当前通知类型(%s)系统不支持",
	})
}
