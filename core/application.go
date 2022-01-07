package core

import (
	"context"
	"distributed-worker/dto"
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

	"github.com/ppkg/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ApplicationContext struct {
	conf       Config
	grpcServer *grpc.Server
	// 调度器master节点
	masterNode dto.NodeInfo
	masterConn *grpc.ClientConn
	lock       sync.Mutex
	// 支持的插件集合
	pluginSet map[string]Plugin
}

func NewApp(opts ...Option) *ApplicationContext {
	instance := &ApplicationContext{}
	instance.initDefaultConfig()
	for _, m := range opts {
		m(&instance.conf)
	}
	instance.initGrpc()
	return instance
}

// 注册插件
func (s *ApplicationContext) RegisterPlugin(plugin Plugin) {
	s.pluginSet[plugin.Name()] = plugin
}

// 初始化默认配置
func (s *ApplicationContext) initDefaultConfig() {
	s.conf.AppName = os.Getenv("APP_NAME")
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

func (s *ApplicationContext) Run() error {
	// 检查调度服务url是否配置
	if s.conf.SchedulerUrl == "" {
		glog.Errorf("Application/run 调度服务url未配置")
		return errors.New("调度服务url未配置")
	}

	// 定时发送心跳并向调度器注册endpoint
	go s.cronHeartbeat()

	// 初始化grpc服务
	err := s.doServe()
	if err != nil {
		glog.Errorf("Application/run 监听grpc服务异常,err:%v", err)
		return err
	}
	return nil
}

// 定时发送心跳
func (s *ApplicationContext) cronHeartbeat() {
	timer := time.Tick(1200 * time.Millisecond)
	var client node.NodeServiceClient
	var err error
	endpoint := fmt.Sprintf("%s:%d", s.conf.Endpoint, s.conf.Port)
	prefix := s.conf.AppName
	if prefix == "" {
		prefix = "worker"
	}
	nodeId := prefix + strings.ReplaceAll(endpoint, ".", "_")
	plugins := make([]string, 0, len(s.pluginSet))
	for k := range s.pluginSet {
		plugins = append(plugins, k)
	}
	for _ = range timer {
		if s.masterConn == nil {
			err = s.initMasterConn()
			if err != nil {
				glog.Errorf("ApplicationContext/cronHeartbeat %v", err)
				continue
			}
		}

		client = node.NewNodeServiceClient(s.masterConn)
		ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
		_, err = client.HeartBeat(ctx, &node.HeartBeatRequest{
			NodeInfo: &node.NodeInfo{
				NodeId:   nodeId,
				Endpoint: endpoint,
			},
			PluginSet: plugins,
		})
		cancel()
		if err == nil {
			continue
		}

		glog.Errorf("ApplicationContext/cronHeartbeat 心跳保持异常,err:%+v", util.ConvertGrpcError(err))
		// 请求调度器master服务出错则关闭连接然后重新创建连接
		s.closeMasterConn()
	}

}

func (s *ApplicationContext) closeMasterConn() {
	if s.masterConn == nil {
		return
	}
	s.masterConn.Close()
	s.masterConn = nil
}

// 初始化调度器master连接
func (s *ApplicationContext) initMasterConn() error {
	s.closeMasterConn()
	// 重置调度器master节点信息
	s.masterNode = dto.NodeInfo{}
	node := s.GetMasterNode()
	var err error
	s.masterConn, err = grpc.Dial(node.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(30*1024*1024)))
	if err != nil {
		return fmt.Errorf("无法打开调度器master节点连接,code:%+v", err)
	}
	return nil
}

// 获取主节点信息
func (s *ApplicationContext) GetMasterNode() dto.NodeInfo {
	if s.masterNode.NodeId != "" {
		return s.masterNode
	}

	retryCount := 1
	for {
		err := s.requestMasterNode()
		if err == nil {
			break
		}
		glog.Errorf("ApplicationContext/GetMasterNode %v", err.Error())
		time.Sleep(3 * time.Second)
		glog.Errorf("ApplicationContext/GetMasterNode 每隔3秒执行第%d次重试", retryCount)
		retryCount++
	}
	return s.masterNode
}

// 请求获取主节点信息
func (s *ApplicationContext) requestMasterNode() error {
	conn, err := grpc.Dial(s.conf.SchedulerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(30*1024*1024)))
	if err != nil {
		return fmt.Errorf("无法打开调度器连接,code:%+v", err)
	}
	defer conn.Close()
	client := node.NewNodeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := client.GetMaster(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("请求调度器master节点信息异常,code:%+v", err)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.masterNode.NodeId = resp.NodeInfo.NodeId
	s.masterNode.Endpoint = resp.NodeInfo.Endpoint
	return nil
}

func (s *ApplicationContext) initGrpc() {
	s.grpcServer = grpc.NewServer()
}

// 注册grpc服务
func (s *ApplicationContext) RegisterGrpc(f func(server *grpc.Server)) *ApplicationContext {
	f(s.grpcServer)
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

// 应用配置
type Config struct {
	// 应用名称
	AppName string
	// 服务端点
	Endpoint string
	// 应用监听端口号
	Port int
	// 调度器地址(域名/ip+端口号)
	SchedulerUrl string
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
