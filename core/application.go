package core

import (
	"context"
	"distributed-worker/dto"
	"distributed-worker/proto/job"
	"distributed-worker/proto/node"
	"distributed-worker/util"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	configClient "github.com/nacos-group/nacos-sdk-go/clients/config_client"
	namingClient "github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/ppkg/glog"
	"github.com/ppkg/kit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ApplicationContext struct {
	conf       Config
	grpcServer *grpc.Server
	// 调度器leader节点
	leaderNode dto.NodeInfo
	leaderConn *grpc.ClientConn
	lock       sync.Mutex
	// 支持的插件集合
	pluginSet     map[string]Plugin
	jobNotifyFunc func(appCtx *ApplicationContext, data dto.JobNotify)

	// nacos服务发现客户端
	namingClient namingClient.INamingClient
	// nacos配置服务客户端
	configClient configClient.IConfigClient
}

func NewApp(opts ...Option) *ApplicationContext {
	instance := &ApplicationContext{
		pluginSet: make(map[string]Plugin),
	}
	instance.initDefaultConfig()
	for _, m := range opts {
		m(&instance.conf)
	}
	instance.initGrpc()
	return instance
}

// 注册插件
func (s *ApplicationContext) RegisterPlugin(plugin Plugin) *ApplicationContext {
	s.pluginSet[plugin.Name()] = plugin
	return s
}

// 获取插件
func (s *ApplicationContext) GetPlugin(name string) Plugin {
	return s.pluginSet[name]
}

// 初始化默认配置
func (s *ApplicationContext) initDefaultConfig() {
	s.conf.AppName = os.Getenv("APP_NAME")
	s.conf.ServiceName = "distributed-workder"
	s.conf.Nacos.WorkerStartupTimeKey = "workerStartupTime"
	s.conf.Endpoint = os.Getenv("ENDPOINT")
	if s.conf.Endpoint == "" {
		s.conf.Endpoint = util.GetLocalIp()
	}
	port := os.Getenv("APP_PORT")
	if port != "" {
		s.conf.Port, _ = strconv.Atoi(port)
	}
	if s.conf.Port == 0 {
		s.conf.Port = 8080
	}

	s.conf.SchedulerUrl = os.Getenv("SCHEDULER_URL")
}

// 服务发现，向nacos注册服务
func (s *ApplicationContext) initNacos() error {
	// 如果nacos未配置则不需要初始化
	if len(s.conf.Nacos.Addrs) == 0 {
		return nil
	}

	var err error
	clientConfig := constant.ClientConfig{
		NamespaceId:         s.conf.Nacos.Namespace,
		TimeoutMs:           2000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		RotateTime:          "24h",
		MaxAge:              3,
		LogLevel:            "info",
	}

	serverConfigs := make([]constant.ServerConfig, 0, len(s.conf.Nacos.Addrs))
	for i, host := range s.conf.Nacos.Addrs {
		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr: host,
			Port:   uint64(s.conf.Nacos.Ports[i]),
		})
	}

	s.configClient, err = clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return fmt.Errorf("当前节点:%s，实例化nacos配置服务客户端异常:%v", s.GetNodeId(), err)
	}

	s.namingClient, err = clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return fmt.Errorf("当前应用:%s，实例化nacos服务发现客户端异常:%v", s.conf.AppName, err)
	}

	success, err := s.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          s.conf.Endpoint,
		Port:        uint64(s.conf.Port),
		ServiceName: s.conf.ServiceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"appName": s.conf.AppName,
			"nodeId":  s.GetNodeId(),
		},
		ClusterName: s.conf.Nacos.ClusterName, // default value is DEFAULT
		GroupName:   s.conf.Nacos.GroupName,   // default value is DEFAULT_GROUP
	})
	if err != nil {
		return fmt.Errorf("当前节点:%s，注册服务发现异常:%v", s.GetNodeId(), err)
	}
	if !success {
		return fmt.Errorf("当前节点:%s，注册服务(%s)失败", s.GetNodeId(), s.getEndpoint())
	}

	// 发布启动时间值变更，驱动调度器更新worker
	success, err = s.configClient.PublishConfig(vo.ConfigParam{
		DataId:  s.conf.Nacos.WorkerStartupTimeKey,
		Group:   s.conf.Nacos.ConfigGroup,
		Content: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("当前节点:%s，发布配置异常:%v", s.GetNodeId(), err)
	}
	if !success {
		return fmt.Errorf("当前节点:%s，发布配置失败", s.GetNodeId())
	}
	return nil
}

func (s *ApplicationContext) Run() error {
	// 检查调度服务url是否配置
	if s.conf.SchedulerUrl == "" && len(s.conf.Nacos.Addrs) == 0 {
		err := errors.New("调度服务地址(SchedulerUrl)或Nacos地址未配置")
		glog.Errorf("Application/run %v", err)
		return err
	}

	err := s.initNacos()
	if err != nil {
		glog.Errorf("Application/run 初始化nacos客户端异常,%v", err)
		return err
	}

	// 订阅job异步通知
	if s.jobNotifyFunc != nil {
		go s.doAsyncNotify()
	}

	// 初始化grpc服务
	err = s.doServe()
	if err != nil {
		glog.Errorf("Application/run 监听grpc服务异常,err:%v", err)
		return err
	}
	return nil
}

// 获取调度器地址
func (s *ApplicationContext) getSchedulerUrl() string {
	scheduleUrl := func() string {
		if s.namingClient == nil {
			return s.conf.SchedulerUrl
		}

		instance, err := s.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
			ServiceName: s.conf.Nacos.SchedulerServiceName,
			Clusters: []string{
				s.conf.Nacos.ClusterName,
			},
			GroupName: s.conf.Nacos.GroupName,
		})
		if err != nil {
			glog.Errorf("ApplicationContext/getSchedulerUrl 从nacos获取调度器地址异常,err:%+v", err)
			return s.conf.SchedulerUrl
		}
		return fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	}()
	if scheduleUrl == "" {
		scheduleUrl = "127.0.0.1:8080"
	}
	return scheduleUrl
}

func (s *ApplicationContext) getEndpoint() string {
	return fmt.Sprintf("%s:%d", s.conf.Endpoint, s.conf.Port)
}

func (s *ApplicationContext) GetNodeId() string {
	prefix := s.conf.AppName
	if prefix == "" {
		prefix = "worker"
	}
	return prefix + strings.ReplaceAll(s.getEndpoint(), ".", "_")
}

// 重置调度器leader节点连接
func (s *ApplicationContext) resetLeaderConn() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.leaderConn != nil {
		s.leaderConn.Close()
		s.leaderConn = nil
	}
	s.leaderNode = dto.NodeInfo{}
}

// 获取调度器leader节点连接
func (s *ApplicationContext) GetLeaderConn() *grpc.ClientConn {
	if s.leaderConn == nil {
		_ = s.GetLeaderNode()
	}
	return s.leaderConn
}

// 获取主节点信息
func (s *ApplicationContext) GetLeaderNode() dto.NodeInfo {
	if s.leaderNode.NodeId != "" {
		return s.leaderNode
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.leaderNode.NodeId != "" {
		return s.leaderNode
	}

	retryCount := 1
	for {
		err := s.requestLeaderNode()
		if err == nil {
			break
		}
		glog.Errorf("ApplicationContext/GetLeaderNode %v", err.Error())
		time.Sleep(3 * time.Second)
		retryCount++
	}
	return s.leaderNode
}

// 请求获取主节点信息
func (s *ApplicationContext) requestLeaderNode() error {
	if s.leaderNode.NodeId != "" {
		return nil
	}
	conn, err := grpc.Dial(s.getSchedulerUrl(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(30*1024*1024)))
	if err != nil {
		return fmt.Errorf("无法打开调度器连接,code:%+v", err)
	}
	defer conn.Close()
	client := node.NewNodeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := client.GetLeader(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("获取调度器leader节点信息异常,code:%+v", err)
	}
	s.leaderNode.NodeId = resp.NodeInfo.NodeId
	s.leaderNode.Endpoint = resp.NodeInfo.Endpoint

	s.leaderConn, err = grpc.Dial(s.leaderNode.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(30*1024*1024)))
	if err != nil {
		return fmt.Errorf("无法打开调度器leader节点连接,code:%+v", err)
	}
	return nil
}

func (s *ApplicationContext) initGrpc() {
	s.grpcServer = grpc.NewServer()
}

// 注册grpc服务
func (s *ApplicationContext) RegisterGrpc(f func(appCtx *ApplicationContext, server *grpc.Server)) *ApplicationContext {
	f(s, s.grpcServer)
	return s
}

func (s *ApplicationContext) doServe() error {
	reflection.Register(s.grpcServer)
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	if err := s.grpcServer.Serve(sock); err != nil {
		return fmt.Errorf("failed to grpc serve: %v", err)
	}
	return nil
}

// 订阅异步通知
func (s *ApplicationContext) SubscribeAsyncNotify(callback func(appCtx *ApplicationContext, data dto.JobNotify)) *ApplicationContext {
	s.jobNotifyFunc = callback
	return s
}

func (s *ApplicationContext) doAsyncNotify() error {
	myDuration := 30 * time.Second
	timer := time.NewTimer(myDuration)
	for {
		func() {
			defer func() {
				<-timer.C
				timer.Reset(myDuration)
			}()

			client := job.NewJobServiceClient(s.GetLeaderConn())
			resp, err := client.AsyncNotify(context.Background(), &job.AsyncNotifyRequest{
				NodeId: s.GetNodeId(),
			})
			if err != nil {
				glog.Errorf("订阅异步job通知异常,scheduler:%s,err:%+v", kit.JsonEncode(s.GetLeaderNode()), err)
				return
			}
			glog.Infof("订阅异步job通知成功,scheduler:%s", kit.JsonEncode(s.GetLeaderNode()))
			for {
				data, err := resp.Recv()
				if err != nil {
					glog.Errorf("接收job异步通知异常,scheduler:%s,err:%+v", kit.JsonEncode(s.GetLeaderNode()), err)
					return
				}
				s.jobNotifyFunc(s, dto.JobNotify{
					Id:     data.Id,
					Name:   data.Name,
					Type:   data.Type,
					Status: data.Status,
					Result: data.Result,
				})
			}
		}()
	}

}

// 应用配置
type Config struct {
	// 应用名称
	AppName string
	// 服务名称
	ServiceName string
	// 服务端点
	Endpoint string
	// 应用监听端口号
	Port int
	// 调度器地址(域名/ip+端口号)
	SchedulerUrl string
	// nacos配置
	Nacos NacosConfig
}

type NacosConfig struct {
	// 调度器服务名称
	SchedulerServiceName string
	Addrs                []string
	Ports                []int
	Namespace            string
	GroupName            string
	ClusterName          string

	// nacos配置服务的配置参数
	// worker工作节点启动时间key
	WorkerStartupTimeKey string
	ConfigGroup          string
}

type Option func(conf *Config)

// 配置应用名称
func WithAppNameOption(name string) Option {
	return func(conf *Config) {
		conf.AppName = name
	}
}

func WithPortOption(port int) Option {
	return func(conf *Config) {
		conf.Port = port
	}
}

func WithEndpointOption(endpoint string) Option {
	return func(conf *Config) {
		conf.Endpoint = endpoint
	}
}

func WithSchedulerUrlOption(schedulerUrl string) Option {
	return func(conf *Config) {
		conf.SchedulerUrl = schedulerUrl
	}
}

// 配置nacos服务地址，格式：域名(ip)+端口号
func WithNacosAddrOption(addr string) Option {
	return func(conf *Config) {
		if addr == "" {
			return
		}
		host, port := parseNacosAddr(addr)
		conf.Nacos.Addrs = append(conf.Nacos.Addrs, host)
		conf.Nacos.Ports = append(conf.Nacos.Ports, port)
	}
}

func WithNacosSchedulerServiceNameOption(serviceName string) Option {
	return func(conf *Config) {
		conf.Nacos.SchedulerServiceName = serviceName
	}
}

func WithNacosGroupNameOption(group string) Option {
	return func(conf *Config) {
		conf.Nacos.GroupName = group
	}
}

func WithNacosClusterNameOption(cluster string) Option {
	return func(conf *Config) {
		conf.Nacos.ClusterName = cluster
	}
}

func WithNacosNamespaceOption(namespace string) Option {
	return func(conf *Config) {
		conf.Nacos.Namespace = namespace
	}
}

// 解析nacos地址
func parseNacosAddr(addr string) (string, int) {
	pathInfo := strings.Split(addr, ":")
	port := 8848
	if len(pathInfo) > 1 {
		tmp, _ := strconv.Atoi(pathInfo[1])
		if tmp > 0 {
			port = tmp
		}
	}
	return pathInfo[0], port
}

func WithNacosConfigGroupOption(group string) Option {
	return func(conf *Config) {
		conf.Nacos.ConfigGroup = group
	}
}
