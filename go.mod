module github.com/xbonlinenet/goup

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.26.0

	go.uber.org/atomic => github.com/uber-go/atomic v1.4.0
	go.uber.org/multierr => github.com/uber-go/multierr v1.1.0
	go.uber.org/zap => github.com/uber-go/zap v1.10.0

	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/net => github.com/golang/net v0.0.0-20190607181551-461777fb6f67
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190610200419-93c9922d18ae
	golang.org/x/text => github.com/golang/text v0.3.2
	google.golang.org/appengine => github.com/golang/appengine v1.6.1
	gopkg.in/fsnotify.v1 => github.com/fsnotify/fsnotify v1.4.2
	gopkg.in/jcmturner/gokrb5.v6 => github.com/jcmturner/gokrb5 v6.0.5+incompatible
)

go 1.12

require (
	github.com/Shopify/sarama v1.23.1
	github.com/ajg/form v1.5.1 // indirect
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/bxcodec/faker/v3 v3.3.0
	github.com/deckarep/golang-set v1.7.1
	github.com/fasthttp-contrib/websocket v0.0.0-20160511215533-1f3b11f56072 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gavv/httpexpect v2.0.0+incompatible
	github.com/gin-gonic/gin v1.4.0
	github.com/go-errors/errors v1.0.1
	github.com/go-redis/cache v6.4.0+incompatible
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/imkira/go-interpol v1.1.0 // indirect
	github.com/jinzhu/gorm v1.9.10
	github.com/json-iterator/go v1.1.6
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/olivere/elastic/v7 v7.0.13
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/client_golang v0.9.3
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cast v1.3.0
	github.com/spf13/viper v1.4.0
	github.com/valyala/fasthttp v1.6.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/xbonlinenet/alter/lib v0.0.0-20190611130810-2fbf77997692
	github.com/xbonlinenet/go_config_center v0.0.0-20190910120241-16614327114b
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	github.com/zsais/go-gin-prometheus v0.0.0-20181030200533-58963fb32f54
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
