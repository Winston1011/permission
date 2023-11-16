package rmq

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

// 规范化 key 并添加前缀生成规范化 header
// 流程：将 key 中的下划线转化为中划线，中划线分割的各段首字母大写，其他字母小写，并添加前缀（如果不存在）
// eg: Long-Key/long_KEy/LONG_KEY => X-Zyb-Mq-Long-Key
func key2Header(key string) string {
	s := http.CanonicalHeaderKey(strings.ReplaceAll(key, "_", "-"))
	// only append header prefix when it doesn't exist.
	if !strings.HasPrefix(s, HeaderPre) {
		s = HeaderPre + s
	}
	return s
}

// 去掉前缀，返回规范化 key
// eg: X-Zyb-Mq-Long-Key => Long-Key
func header2Key(key string) string {
	s := http.CanonicalHeaderKey(strings.ReplaceAll(key, "_", "-"))
	return strings.TrimPrefix(s, HeaderPre)
}

func fmtHeaders(m *primitive.Message, prefix string) string {
	headerStr := ""
	for key, value := range m.GetProperties() {
		if strings.HasPrefix(key, prefix) {
			headerStr += fmt.Sprintf("%s:%s;", key, value)
		}
	}
	return headerStr
}

func getHostListByDns(nameServers []string) (hostList []string) {
	for _, ns := range nameServers {
		host, port, err := net.SplitHostPort(ns)
		if err != nil {
			logger.Warn("invalid nameserver config",
				fields(zlog.String("ns", ns), zlog.String("err", err.Error()))...)
			continue
		}
		// have to resolve the domain name to ips
		addrs, err := net.LookupHost(host)
		if err != nil {
			logger.Warn("failed to lookup nameserver",
				fields(zlog.String("ns", ns), zlog.String("err", err.Error()))...)
			continue
		}

		for _, addr := range addrs {
			hostList = append(hostList, addr+":"+port)
		}
	}

	return hostList
}

const (
	packKeyProvider       = "_provider"
	packKeyProduct        = "_product"
	packKeyTopic          = "_topic"
	packKeyCmd            = "_cmd"
	packKeyTransID        = "_transid"
	packKeyLogID          = "_log_id"
	packKeyCallerURI      = "_caller_uri"
	packKeyCommitTime     = "_commit_time"
	packKeyCommitTimeUs   = "_commit_time_us"
	packKeyClientIP       = "_client_ip"
	packKeyIDC            = "_idc"
	packKeyTopicGroupName = "_topic_group_name"
	packKeyCluster        = "_cluster"
)

func formatNmqMsg(ctx *gin.Context, cmd string, topic string, product string, data map[string]interface{}) (mcpack.V2Map, uint64) {

	pack := make(mcpack.V2Map)

	transID := uint64(generateSnowflake())
	pack[packKeyTransID] = transID

	var logID uint32
	if zlog.GetLogID(ctx) != "" {
		var logIDParsed uint64
		logIDParsed, err := strconv.ParseUint(zlog.GetLogID(ctx), 10, 32)
		if err != nil { //兼容非32位uint格式的logid
			logIDParsed = 0
		}
		logID = uint32(logIDParsed)
	}
	pack[packKeyLogID] = logID

	patchCallerURI := ""
	for k, v := range data {
		pack[k] = v
		if k == packKeyCallerURI {
			patchCallerURI, _ = v.(string)
		}
	}
	callerURI, _ := utils.GetPressureFlag(ctx)
	if callerURI == "" {
		// patch: 一个兜底，避免业务不合理的使用了ctx，导致从ctx中获取callerURI失败情况
		if patchCallerURI != "" {
			callerURI = patchCallerURI
		} else {
			callerURI = "\x00"
		}
	}
	pack[packKeyCallerURI] = callerURI

	pack[packKeyProduct] = product
	pack[packKeyTopic] = topic
	pack[packKeyCmd] = cmd

	pack[packKeyProvider] = "RAL"
	currTs := time.Now()
	pack[packKeyCommitTimeUs] = uint32((currTs.UnixNano() - currTs.Unix()*int64(time.Second)) / int64(time.Microsecond))
	pack[packKeyCommitTime] = uint32(currTs.Unix())
	pack[packKeyTopicGroupName] = topic
	pack[packKeyCluster] = "RMQ"
	pack[packKeyIDC] = "cn"
	pack[packKeyClientIP] = env.LocalIP

	return pack, transID
}

var _node *snowflake.Node

func init() {
	// first try get env config
	id := getIDBasedOnEnviron()
	if id == 0 {
		// then try get id based on ip
		id = binary.BigEndian.Uint16(net.ParseIP(env.LocalIP).To4()[2:])
	}

	snowflake.NodeBits = 16 // to hold the last half of the IP, eg. 172.29.240.120 -> 240.120 -> 61560
	snowflake.StepBits = 6  // 6bit for 64 IDs per millisecond, yielding 64000 qps, should be enough for nmqproxy

	var err error
	_node, err = snowflake.NewNode(int64(id))
	if err != nil {
		logger.Error("failed to set worker id", fields()...)
	}
}

func getIDBasedOnEnviron() uint16 {
	if os.Getenv("SNOWFLAKE_ID") != "" {
		wid, err := strconv.ParseUint(os.Getenv("SNOWFLAKE_ID"), 10, 16)
		if err != nil {
			return 0
		}

		return uint16(wid)
	}

	return 0
}

func generateSnowflake() int64 {
	return _node.Generate().Int64()
}
