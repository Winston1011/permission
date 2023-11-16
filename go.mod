module permission

go 1.18

replace (
	github.com/apache/rocketmq-client-go/v2 => ./pkg/rocketmq-client-go
	github.com/gin-gonic/gin => ./pkg/gin
	github.com/olivere/elastic/v7 => ./pkg/elastic
)

require (
	git.apache.org/thrift.git v0.13.0
	github.com/Shopify/sarama v1.28.0
	github.com/apache/rocketmq-client-go/v2 v2.1.0-rc5
	github.com/bwmarrin/snowflake v0.3.0
	github.com/felixge/fgprof v0.9.3
	github.com/gin-gonic/gin v1.9.1
	github.com/go-sql-driver/mysql v1.7.0
	github.com/golang/protobuf v1.5.3
	github.com/gomodule/redigo v1.8.9
	github.com/google/uuid v1.3.0
	github.com/json-iterator/go v1.1.12
	github.com/mitchellh/mapstructure v1.5.0
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/olivere/elastic/v7 v7.0.31
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.17.0
	github.com/shirou/gopsutil/v3 v3.23.10
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.4
	go.uber.org/automaxprocs v1.5.3
	go.uber.org/zap v1.24.0
	golang.org/x/net v0.10.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.5.2
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5

)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.3.4 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20220526154910-8bf9453eb81a // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.11.2 // indirect
	github.com/goccy/go-json v0.9.10 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.11.7 // indirect
	github.com/klauspost/cpuid/v2 v2.0.14 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-sqlite3 v1.14.17 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/sirupsen/logrus v1.4.1 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tidwall/gjson v1.13.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.9 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/arch v0.0.0-20220412001346-fc48f9fe4c15 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	stathat.com/c/consistent v1.0.0 // indirect
)
