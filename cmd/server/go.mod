module github.com/uber/cadence/cmd/server

go 1.20

// build against the current code in the "main" (and gcloud) module, not a specific SHA.
//
// anyone outside this repo using this needs to ensure that both the "main" module and this module
// are at the same SHA for consistency, but internally we can cheat by telling Go that it's at a
// relative file path.
replace github.com/uber/cadence => ../..

replace github.com/uber/cadence/common/archiver/gcloud => ../../common/archiver/gcloud

require (
	github.com/Shopify/sarama v1.33.0 // indirect
	github.com/VividCortex/mysqlerr v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.180 // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20191106001114-12b4e2b38748 // indirect
	github.com/cch123/elasticsql v0.0.0-20190321073543-a1a440758eb9 // indirect
	github.com/cristalhq/jwt/v3 v3.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gocql/gocql v1.7.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.5.0 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/iancoleman/strcase v0.2.0 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20200615141059-0794cb1f47ee // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/olivere/elastic v6.2.37+incompatible // indirect
	github.com/olivere/elastic/v7 v7.0.21 // indirect
	github.com/opensearch-project/opensearch-go/v2 v2.2.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pborman/uuid v0.0.0-20180906182336-adf5a7427709 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/startreedata/pinot-client-go v0.2.0 // latest release supports pinot v0.12.0 which is also internal version
	github.com/stretchr/testify v1.9.0
	github.com/uber-go/tally v3.3.15+incompatible // indirect
	github.com/uber/cadence-idl v0.0.0-20241126065313-57bd6876d48f
	github.com/uber/ringpop-go v0.8.5 // indirect
	github.com/uber/tchannel-go v1.22.2 // indirect
	github.com/valyala/fastjson v1.4.1 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/cadence v0.19.0
	go.uber.org/config v1.4.0 // indirect
	go.uber.org/fx v1.13.1 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/thriftrw v1.29.2 // indirect
	go.uber.org/yarpc v1.70.3 // indirect
	go.uber.org/zap v1.13.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.16.0 // indirect
	gonum.org/v1/gonum v0.7.0 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	gopkg.in/validator.v2 v2.0.0-20180514200540-135c24b11c19 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

require (
	github.com/uber/cadence v0.0.0-00010101000000-000000000000
	github.com/uber/cadence/common/archiver/gcloud v0.0.0-00010101000000-000000000000
)

require (
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
)

require (
	cloud.google.com/go v0.110.8 // indirect
	cloud.google.com/go/compute v1.23.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.3 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/MicahParks/keyfunc/v2 v2.1.0 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20161002113705-648efa622239 // indirect
	github.com/apache/thrift v0.17.0 // indirect
	github.com/benbjohnson/clock v0.0.0-20161215174838-7dc76406b6d3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emirpasic/gods v0.0.0-20190624094223-e689965507ab // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/go-zookeeper/zk v1.0.3 // indirect
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/gogo/status v1.1.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.4 // indirect
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
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kisielk/errcheck v1.5.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.1 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/uber-common/bark v1.2.1 // indirect
	github.com/uber-go/mapdecode v1.0.0 // indirect
	github.com/urfave/cli/v2 v2.27.4
	github.com/xdg/stringprep v1.0.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.uber.org/dig v1.10.0 // indirect
	go.uber.org/net/metrics v1.3.0 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/exp v0.0.0-20231226003508-02704c960a9b // indirect
	golang.org/x/exp/typeparams v0.0.0-20220218215828-6cf2b201936e // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/oauth2 v0.11.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.128.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231012201019-e917dd12ba7a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231030173426-d783a09b4405 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.3.2 // indirect
)
