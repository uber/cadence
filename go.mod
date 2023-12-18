module github.com/uber/cadence

go 1.20

require (
	cloud.google.com/go/storage v1.30.1
	github.com/Shopify/sarama v1.33.0
	github.com/VividCortex/mysqlerr v1.0.0
	github.com/aws/aws-sdk-go v1.44.180
	github.com/cactus/go-statsd-client/statsd v0.0.0-20191106001114-12b4e2b38748
	github.com/cch123/elasticsql v0.0.0-20190321073543-a1a440758eb9
	github.com/cristalhq/jwt/v3 v3.1.0
	github.com/dave/dst v0.26.2
	github.com/davecgh/go-spew v1.1.1
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/emirpasic/gods v1.12.0
	github.com/fatih/color v1.13.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/gocql/gocql v0.0.0-20211015133455-b225f9b53fa1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.5.0
	github.com/hashicorp/go-version v1.2.0
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/jmoiron/sqlx v1.2.1-0.20200615141059-0794cb1f47ee
	github.com/jonboulle/clockwork v0.1.0
	github.com/lib/pq v1.2.0
	github.com/m3db/prometheus_client_golang v0.8.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/olivere/elastic v6.2.37+incompatible
	github.com/olivere/elastic/v7 v7.0.21
	github.com/opensearch-project/opensearch-go/v2 v2.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/otiai10/copy v1.1.1
	github.com/pborman/uuid v1.2.1
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/startreedata/pinot-client-go v0.0.0-20230303070132-3b84c28a9e95 // latest release doesn't support pinot v0.12, so use master branch
	github.com/stretchr/testify v1.8.4
	github.com/uber-go/tally v3.5.8+incompatible
	github.com/uber/cadence-idl v0.0.0-20230905165949-03586319b849
	github.com/uber/ringpop-go v0.8.5
	github.com/uber/tchannel-go v1.34.1
	github.com/urfave/cli v1.22.4
	github.com/valyala/fastjson v1.4.1
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2
	go.mongodb.org/mongo-driver v1.7.3
	go.uber.org/atomic v1.11.0
	go.uber.org/cadence v1.2.7
	go.uber.org/config v1.4.0
	go.uber.org/fx v1.13.1
	go.uber.org/multierr v1.11.0
	go.uber.org/thriftrw v1.31.0
	go.uber.org/yarpc v1.71.0
	go.uber.org/zap v1.26.0
	golang.org/x/net v0.19.0
	golang.org/x/oauth2 v0.13.0
	golang.org/x/sync v0.5.0
	golang.org/x/time v0.5.0
	golang.org/x/tools v0.16.1
	gonum.org/v1/gonum v0.14.0
	google.golang.org/api v0.149.0
	google.golang.org/grpc v1.60.0
	gopkg.in/validator.v2 v2.0.0-20180514200540-135c24b11c19
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/google/go-cmp v0.6.0

require (
	cloud.google.com/go v0.110.10 // indirect
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.5 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/apache/thrift v0.19.0 // indirect
	github.com/benbjohnson/clock v0.0.0-20161215174838-7dc76406b6d3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/gogo/status v1.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/marusama/semaphore/v2 v2.5.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/stretchr/objx v0.5.1 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/uber-common/bark v1.2.1 // indirect
	github.com/uber-go/mapdecode v1.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/dig v1.10.0 // indirect
	go.uber.org/net/metrics v1.4.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231214170342-aacd6d4b4611 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20231211222908-989df2bf70f3 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.4.6 // indirect
)

// ringpop-go and tchannel-go depends on older version of thrift, yarpc brings up newer version
replace github.com/apache/thrift => github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7

// DO NOT USE as it misses mysql/config store fix
retract v1.2.3
