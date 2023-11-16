package zns

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"gopkg.in/yaml.v2"
	"permission/pkg/zns/bns"
	"permission/pkg/zns/logger"
	"permission/pkg/zns/util"
)

type Instance struct {
	IP   string
	Port string
}

// 根据 znsName 查询有效ip:port
func GetZnsInstance(ctx *gin.Context, service, name string) (list []*Instance, err error) {
	if v, ok := ins.Load(service); ok {
		if list, ok = v.([]*Instance); ok {
			return list, nil
		}
	}

	if name != "" {
		list, err = download(ctx, service, name)
	}

	return list, err
}

func download(ctx *gin.Context, service, znsName string) ([]*Instance, error) {
	var list []*Instance
	insInfoList, err := resolve(ctx, znsName)
	if err != nil {
		znsLogger.Logger(ctx, logger.Warn, "resolve failed, service: "+service+"  zns: "+znsName)
		return nil, err
	}

	for _, value := range insInfoList {
		if *value.InstanceStatus.Status != 0 {
			continue
		}

		in := &Instance{
			IP:   util.UInt32IpToString(*value.HostIp),
			Port: strconv.Itoa(int(*value.InstanceStatus.Port)),
		}
		list = append(list, in)
		znsLogger.Logger(ctx, logger.Debug, "resolve success, zns: "+znsName+" "+in.IP+":"+in.Port)
	}

	if service != "" {
		ins.Delete(service)
		ins.Store(service, list)
	}

	return nil, nil
}

func resolve(ctx *gin.Context, name string) ([]*bns.InstanceInfo, error) {
	bnsClient := bns.New(config.LocalAddr, config.Timeout)
	if err := bnsClient.Connect(); err != nil {
		znsLogger.Logger(ctx, logger.Error, "zns connect error:"+err.Error())
		return nil, err
	}
	defer bnsClient.Close()

	req := &bns.LocalNamingRequest{
		ServiceName: proto.String(name),
		All:         proto.Bool(true),
		Type:        proto.Int32(0),
	}
	var rsp bns.LocalNamingResponse
	if err := bnsClient.Call(req, &rsp); err != nil {
		znsLogger.Logger(ctx, logger.Error, "call zns error: "+err.Error()+" name:")
		return nil, err
	}

	return rsp.InstanceInfo, nil
}

func refresh() {
	defer func() {
		if err := recover(); err != nil {
			znsLogger.Logger(nil, logger.Error, fmt.Sprintf("zns download panic, error: %+v", err))
			refresh()
		}
	}()

	for {
		for service, zns := range znsList {
			_, _ = download(nil, service, zns)
		}
		time.Sleep(config.Interval)
	}
}

var znsLogger logger.Interface

type Config struct {
	Logger     logger.Interface
	Path       string
	Timeout    time.Duration `yaml:"timeout"`
	Interval   time.Duration `yaml:"interval"`
	LocalAddr  string        `yaml:"localAddr"`
	RemoteAddr string        `yaml:"remoteAddr"`
}

func (conf *Config) checkConf() {
	if conf.Timeout == 0 {
		conf.Timeout = 5 * time.Second
	}

	if conf.Interval == 0 {
		conf.Interval = 10 * time.Second
	}

	if conf.LocalAddr == "" {
		conf.LocalAddr = "localhost:793"
		conf.RemoteAddr = "localhost:793"
	} else {
		conf.RemoteAddr = conf.LocalAddr
	}
}

// 支持重写配置
func SetConf(conf *Config) {
	config = conf
}

var (
	// zns 全局配置
	config *Config
	// 需要解析的zns列表
	znsList map[string]string
	// zns 对应的实例
	ins sync.Map
)

func Init(conf *Config) {
	conf.checkConf()
	if conf.Logger == nil {
		znsLogger = logger.Default
	} else {
		znsLogger = conf.Logger
	}
	config = conf

	// 获得需要解析的zns列表
	znsList = make(map[string]string)
	if yamlFile, err := ioutil.ReadFile(config.Path); err != nil {
		panic(" get error: %v " + err.Error())
	} else if err = yaml.Unmarshal(yamlFile, znsList); err != nil {
		panic(" unmarshal error: %v" + err.Error())
	}

	if len(znsList) == 0 {
		return
	}

	znsLogger.Logger(nil, logger.Info, "[prot=zns] start to resolve zns background")

	go refresh()
}
