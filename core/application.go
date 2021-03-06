package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ppkg/distributed-worker/dto"
	"github.com/ppkg/distributed-worker/enum"
	"github.com/ppkg/distributed-worker/proto/job"
	"github.com/ppkg/distributed-worker/proto/node"
	"github.com/ppkg/distributed-worker/proto/task"
	"github.com/ppkg/distributed-worker/util"
	"github.com/ppkg/kit"

	"github.com/maybgit/glog"
	"github.com/nacos-group/nacos-sdk-go/clients"
	configClient "github.com/nacos-group/nacos-sdk-go/clients/config_client"
	namingClient "github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	nacosModel "github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ApplicationContext struct {
	conf       Config
	grpcServer *grpc.Server
	isInit     bool
	// 调度器leader节点
	leaderEndpoint string
	leaderConn     *grpc.ClientConn
	lock           sync.Mutex
	// 支持的插件集合
	pluginSet map[string]PluginHandler
	// job回调通知handler集合
	jobNotifySet map[string]JobNotifyHandler

	// nacos服务发现客户端
	namingClient namingClient.INamingClient
	// nacos配置服务客户端
	configClient configClient.IConfigClient
}

func NewApp(opts ...Option) *ApplicationContext {
	instance := &ApplicationContext{
		pluginSet:    make(map[string]PluginHandler),
		jobNotifySet: make(map[string]JobNotifyHandler),
	}
	instance.initDefaultConfig()
	for _, m := range opts {
		m(&instance.conf)
	}
	instance.initGrpc()
	return instance
}

// 注册插件
func (s *ApplicationContext) RegisterPlugin(fn func(ctx *ApplicationContext) PluginHandler) *ApplicationContext {
	handler := fn(s)
	s.pluginSet[handler.Name()] = handler
	return s
}

// 获取插件handler
func (s *ApplicationContext) GetPluginHandler(name string) PluginHandler {
	return s.pluginSet[name]
}

// 注册job回调通知
func (s *ApplicationContext) RegisterJobNotify(fn func(ctx *ApplicationContext) JobNotifyHandler) *ApplicationContext {
	handler := fn(s)
	s.jobNotifySet[handler.Name()] = handler
	return s
}

// 获取job回调通知handler
func (s *ApplicationContext) GetJobNotifyHandler(name string) JobNotifyHandler {
	return s.jobNotifySet[name]
}

// 获取所有已注册插件名称
func (s *ApplicationContext) GetPluginNameSet() []string {
	list := make([]string, 0, len(s.pluginSet))
	for name := range s.pluginSet {
		list = append(list, name)
	}
	return list
}

// 获取所有已注册job通知处理器名称
func (s *ApplicationContext) GetJobNotifyNameSet() []string {
	list := make([]string, 0, len(s.jobNotifySet))
	for name := range s.jobNotifySet {
		list = append(list, name)
	}
	return list
}

// 初始化默认配置
func (s *ApplicationContext) initDefaultConfig() {
	s.conf.AppName = os.Getenv("APP_NAME")
	if s.conf.AppName == "" {
		s.conf.AppName = "distributed-workder"
	}
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

	nacosAddrs := os.Getenv("NACOS_ADDRS")
	if nacosAddrs != "" {
		addrList := strings.Split(nacosAddrs, ",")
		for _, item := range addrList {
			s.appendNacosAddrConfig(item)
		}
	}

	s.conf.SchedulerServiceName = os.Getenv("SCHEDULER_SERVICE_NAME")
	if s.conf.SchedulerServiceName == "" {
		s.conf.SchedulerServiceName = "distributed-scheduler"
	}

	s.conf.Nacos.ClusterName = os.Getenv("NACOS_CLUSTER_NAME")
	s.conf.Nacos.ServiceGroup = os.Getenv("NACOS_SERVICE_GROUP")
	if s.conf.Nacos.ServiceGroup == "" {
		s.conf.Nacos.ServiceGroup = "DEFAULT_GROUP"
	}
	s.conf.Nacos.Namespace = os.Getenv("NACOS_NAMESPACE")

	s.conf.WorkerServiceName = "distributed-workder"
	s.conf.LeaderStateKey = "leaderState"
}

func (s *ApplicationContext) appendNacosAddrConfig(addr string) {
	if addr == "" {
		return
	}
	host, port := parseNacosAddr(addr)
	s.conf.Nacos.Addrs = append(s.conf.Nacos.Addrs, host)
	s.conf.Nacos.Ports = append(s.conf.Nacos.Ports, port)
}

// 服务发现，向nacos注册服务
func (s *ApplicationContext) initNacos() error {
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
	return nil
}

// 注册服务发现
func (s *ApplicationContext) initNacosDiscovery() error {
	success, err := s.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          s.conf.Endpoint,
		Port:        uint64(s.conf.Port),
		ServiceName: s.conf.WorkerServiceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata: map[string]string{
			"appName":      s.conf.AppName,
			"nodeId":       s.GetNodeId(),
			"pluginSet":    strings.Join(s.GetPluginNameSet(), ","),
			"jobNotifySet": strings.Join(s.GetJobNotifyNameSet(), ","),
		},
		ClusterName: s.conf.Nacos.ClusterName,  // default value is DEFAULT
		GroupName:   s.conf.Nacos.ServiceGroup, // default value is DEFAULT_GROUP
	})
	if err != nil {
		return fmt.Errorf("当前节点:%s，注册服务发现异常:%v", s.GetNodeId(), err)
	}
	if !success {
		return fmt.Errorf("当前节点:%s，注册服务(%s)失败", s.GetNodeId(), s.getEndpoint())
	}
	return nil
}

// 初始化处理
func (s *ApplicationContext) Init() error {
	if s.isInit {
		return nil
	}
	// 检查nacos服务是否已配置
	if len(s.conf.Nacos.Addrs) == 0 {
		err := errors.New("Nacos服务地址未配置")
		glog.Errorf("ApplicationContext/Init %v", err)
		return err
	}

	err := s.initNacos()
	if err != nil {
		glog.Errorf("ApplicationContext/Init 初始化nacos客户端异常,%v", err)
		return err
	}

	// 监听调度器服务发现变更
	err = s.watchSchedulerService()
	if err != nil {
		glog.Errorf("ApplicationContext/run 监控调度器服务异常,%v", err)
		return err
	}

	// 监听调度器状态
	err = s.watchRaftLeaderState()
	if err != nil {
		glog.Errorf("ApplicationContext/run 监控调度器状态异常,%v", err)
		return err
	}

	s.isInit = true
	return nil
}

func (s *ApplicationContext) Run() error {
	err := s.Init()
	if err != nil {
		return err
	}

	// 注册服务发现
	err = s.initNacosDiscovery()
	if err != nil {
		glog.Errorf("ApplicationContext/run 注册服务发现异常,%v", err)
		return err
	}

	// 注册内置grpc服务
	s.RegisterGrpc(func(appCtx *ApplicationContext, server *grpc.Server) {
		job.RegisterJobServiceServer(server, NewJobService(appCtx))
	})
	s.RegisterGrpc(func(appCtx *ApplicationContext, server *grpc.Server) {
		task.RegisterTaskServiceServer(server, NewTaskService(appCtx))
	})

	glog.Infof("worker(%s)已启动,endpoint地址:%s", s.conf.AppName, s.getEndpoint())
	// 初始化grpc服务
	err = s.doServe()
	if err != nil {
		glog.Errorf("ApplicationContext/run 监听grpc服务异常,err:%v", err)
		return err
	}
	return nil
}

// 监听scheduler服务发现
func (s *ApplicationContext) watchSchedulerService() error {
	return s.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: s.conf.SchedulerServiceName,
		GroupName:   s.conf.Nacos.ServiceGroup,
		Clusters: []string{
			s.conf.Nacos.ClusterName,
		},
		SubscribeCallback: func(services []nacosModel.SubscribeService, nacosErr error) {
			glog.Infof("ApplicationContext/watchSchedulerService 收到scheduler服务发现通知:%s,%v", kit.JsonEncode(services), nacosErr)
			s.dynamicUpdateScheduler()
		},
	})
}

// 监控调度器(raft leader)状态变更
func (s *ApplicationContext) watchRaftLeaderState() error {
	return s.configClient.ListenConfig(vo.ConfigParam{
		DataId: s.conf.LeaderStateKey,
		Group:  s.conf.Nacos.ServiceGroup,
		OnChange: func(namespace, group, dataId, data string) {
			glog.Infof("ApplicationContext/watchRaftLeaderState 收到leader变更通知:%s,%s,%s,%s", namespace, group, dataId, data)
			s.dynamicUpdateScheduler()
		},
	})
}

// 动态更新调度器
func (s *ApplicationContext) dynamicUpdateScheduler() {
	// 对于已重置的调度器连接直接return
	if s.leaderEndpoint == "" {
		return
	}

	serviceList := s.getServiceList(s.conf.SchedulerServiceName)
	if len(serviceList) == 0 {
		return
	}
	glog.Infof("ApplicationContext/dynamicUpdateScheduler scheduler健康实例:%s", kit.JsonEncode(serviceList))
	var leaderIntance nacosModel.Instance
	isFound := false
	for _, item := range serviceList {
		if enum.RaftRole(item.Metadata["role"]) == enum.LeaderRaftRole {
			leaderIntance = item
			isFound = true
			break
		}
	}
	if !isFound {
		return
	}

	// 如果是端点相等则说明调度器leader没有变
	if leaderIntance.Metadata["endpoint"] == s.leaderEndpoint {
		return
	}

	glog.Infof("ApplicationContext/dynamicUpdateScheduler 当前节点:%s,raft集群leader节点由%s变更为%s", s.GetNodeId(), s.leaderEndpoint, leaderIntance.Metadata["endpoint"])
	// 如果端点不相等则说明调度器leader有变化，需要重置leader连接
	s.resetLeaderConn()
}

// 重置调度器leader连接
func (s *ApplicationContext) resetLeaderConn() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.leaderEndpoint = ""
	if s.leaderConn != nil {
		_ = s.leaderConn.Close()
		s.leaderConn = nil
	}
}

// 获取健康的服务列表
func (s *ApplicationContext) getServiceList(serviceName string) []nacosModel.Instance {
	instanceList, err := s.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
		Clusters: []string{
			s.conf.Nacos.ClusterName,
		},
		GroupName: s.conf.Nacos.ServiceGroup,
	})
	if err == nil {
		return instanceList
	}

	if strings.Contains(err.Error(), "instance list is empty!") {
		return nil
	}
	glog.Errorf("ApplicationContext/getServiceList 从nacos中获取服务列表(%s)异常,err:%+v", serviceName, err)
	return nil
}

// 获取调度器地址
func (s *ApplicationContext) getSchedulerUrl() string {
	instance, err := s.namingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: s.conf.SchedulerServiceName,
		Clusters: []string{
			s.conf.Nacos.ClusterName,
		},
		GroupName: s.conf.Nacos.ServiceGroup,
	})
	if err == nil {
		return fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
	}
	glog.Errorf("ApplicationContext/getSchedulerUrl 从nacos获取调度器地址异常,err:%+v", err)
	return "127.0.0.1:8080"
}

func (s *ApplicationContext) getEndpoint() string {
	return fmt.Sprintf("%s:%d", s.conf.Endpoint, s.conf.Port)
}

func (s *ApplicationContext) GetNodeId() string {
	prefix := s.conf.AppName
	if prefix == "" {
		prefix = "worker"
	}
	return prefix + "_" + strings.ReplaceAll(strings.ReplaceAll(s.getEndpoint(), ":", "_"), ".", "_")
}

// 获取调度器leader节点连接
func (s *ApplicationContext) GetLeaderConn() *grpc.ClientConn {
	if s.leaderConn == nil {
		_ = s.GetLeaderEndpoint()
	}
	return s.leaderConn
}

// 获取主节点端点
func (s *ApplicationContext) GetLeaderEndpoint() string {
	if s.leaderEndpoint != "" {
		return s.leaderEndpoint
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.leaderEndpoint != "" {
		return s.leaderEndpoint
	}

	retryCount := 1
	for {
		err := s.requestLeaderNode()
		if err == nil {
			break
		}
		glog.Errorf("ApplicationContext/GetLeaderEndpoint 第%d次重试,%v", retryCount, err.Error())
		time.Sleep(3 * time.Second)
		retryCount++
	}
	return s.leaderEndpoint
}

// 请求获取主节点信息
func (s *ApplicationContext) requestLeaderNode() error {
	maxSize := 50 * 1024 * 1024
	conn, err := grpc.Dial(s.getSchedulerUrl(), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	s.leaderEndpoint = resp.NodeInfo.Endpoint

	s.leaderConn, err = grpc.Dial(s.leaderEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxSize),
		grpc.MaxCallSendMsgSize(maxSize),
	))
	if err != nil {
		return fmt.Errorf("无法打开调度器leader节点连接,code:%+v", err)
	}
	return nil
}

func (s *ApplicationContext) initGrpc() {
	maxSize := 50 * 1024 * 1024
	s.grpcServer = grpc.NewServer(grpc.MaxRecvMsgSize(maxSize),
		grpc.MaxSendMsgSize(maxSize))
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

// 同步提交job请求(当调度器挂掉会导致job处理中断，谨用)
func (s *ApplicationContext) SyncSubmitJob(req dto.SyncJobRequest) (resp dto.SyncJobResponse, err error) {
	client := job.NewJobServiceClient(s.GetLeaderConn())
	stream, err := client.SyncSubmit(context.Background())
	if err != nil {
		glog.Errorf("ApplicationContext/SyncSubmitJob 发起同步job请求异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
		return
	}

	for _, item := range req.TaskInputList {
		err = stream.Send(&job.SyncSubmitRequest{
			Name:                   req.Name,
			Meta:                   req.Meta,
			Label:                  req.Label,
			TaskExceptionOperation: int32(req.TaskExceptionOperation),
			Source:                 s.conf.AppName,
			PluginSet:              req.PluginSet,
			Data:                   item,
		})
		if err != nil {
			glog.Errorf("ApplicationContext/SyncSubmitJob 发送同步job数据异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
			return
		}
	}

	rpcResp, err := stream.CloseAndRecv()
	if err != nil {
		glog.Errorf("ApplicationContext/SyncSubmitJob 接收同步job数据异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
		return
	}

	resp.Id = rpcResp.Id
	resp.Meta = rpcResp.Meta
	resp.Status = enum.JobStatus(rpcResp.Status)
	resp.Result = rpcResp.Result
	resp.Message = rpcResp.Mesage
	return
}

// 异步提交job请求
func (s *ApplicationContext) AsyncSubmitJob(req dto.AsyncJobRequest) (jobId int64, err error) {
	client := job.NewJobServiceClient(s.GetLeaderConn())
	stream, err := client.AsyncSubmit(context.Background())
	if err != nil {
		glog.Errorf("ApplicationContext/AsyncSubmitJob 发起异步job请求异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
		return
	}

	for _, item := range req.TaskInputList {
		err = stream.Send(&job.AsyncSubmitRequest{
			Name:                   req.Name,
			Label:                  req.Label,
			Meta:                   req.Meta,
			TaskExceptionOperation: int32(req.TaskExceptionOperation),
			Source:                 s.conf.AppName,
			Type:                   req.Type,
			IsNotify:               req.IsNotify,
			PluginSet:              req.PluginSet,
			Data:                   item,
		})
		if err != nil {
			glog.Errorf("ApplicationContext/AsyncSubmitJob 发送异步job数据异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
			return
		}
	}

	rpcResp, err := stream.CloseAndRecv()
	if err != nil {
		glog.Errorf("ApplicationContext/AsyncSubmitJob 接收异步job数据异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
		return
	}
	return rpcResp.Id, nil
}

// 应用配置
type Config struct {
	// 应用名称
	AppName string
	// 服务端点
	Endpoint string
	// 应用监听端口号
	Port int
	// 调度器服务名称
	SchedulerServiceName string
	// worker服务名称
	WorkerServiceName string
	// raft leader配置中心key
	LeaderStateKey string
	// nacos配置
	Nacos NacosConfig
}

type NacosConfig struct {
	Addrs        []string
	Ports        []int
	Namespace    string
	ServiceGroup string
	ClusterName  string
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

func WithSchedulerServiceNameOption(serviceName string) Option {
	return func(conf *Config) {
		conf.SchedulerServiceName = serviceName
	}
}

func WithNacosServiceGroupOption(group string) Option {
	return func(conf *Config) {
		conf.Nacos.ServiceGroup = group
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

// 手动取消job
func (s *ApplicationContext) ManualCancelJob(req dto.ManualCancelRequest) error {
	if req.Id == 0 {
		return fmt.Errorf("参数ID不能为空")
	}
	client := job.NewJobServiceClient(s.GetLeaderConn())
	resp, err := client.ManualCancel(context.Background(), &job.ManualCancelRequest{
		Id:     req.Id,
		Reason: req.Reason,
	})
	if err != nil {
		glog.Errorf("ApplicationContext/ManualCancelJob 发起手动取消job请求异常,参数:%s,err:%+v", kit.JsonEncode(req), err)
		return err
	}
	glog.Infof("ApplicationContext/ManualCancelJob 手动取消job返回数据:%s,请求参数:%s", kit.JsonEncode(resp), kit.JsonEncode(req))
	if resp.Status == int32(enum.FailCancelStatus) {
		return errors.New(resp.Message)
	}
	return nil
}
