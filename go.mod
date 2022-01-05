module github.com/uber/cadence

go 1.12

require (
	cloud.google.com/go/bigquery v1.6.0 // indirect
	cloud.google.com/go/storage v1.6.0
	github.com/DataDog/zstd v1.4.0 // indirect
	github.com/Shopify/sarama v1.23.0
	github.com/VividCortex/mysqlerr v1.0.0
	github.com/apache/thrift v0.13.0
	github.com/aws/aws-sdk-go v1.34.13
	github.com/benbjohnson/clock v0.0.0-20161215174838-7dc76406b6d3 // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20191106001114-12b4e2b38748
	github.com/cch123/elasticsql v0.0.0-20190321073543-a1a440758eb9
	github.com/cristalhq/jwt/v3 v3.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13
	github.com/dmarkham/enumer v1.5.1
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/emirpasic/gods v0.0.0-20190624094223-e689965507ab
	github.com/fatih/color v1.10.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocql/gocql v0.0.0-20211015133455-b225f9b53fa1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/uuid v1.1.2
	github.com/hashicorp/go-version v1.2.0
	github.com/iancoleman/strcase v0.0.0-20190422225806-e506e3ef7365
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20200615141059-0794cb1f47ee
	github.com/jonboulle/clockwork v0.1.0
	github.com/lib/pq v1.2.0
	github.com/m3db/prometheus_client_golang v0.8.1
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/mattn/go-sqlite3 v1.11.0 // indirect
	github.com/mattn/goveralls v0.0.7
	github.com/mgechev/revive v1.0.3
	github.com/olekukonko/tablewriter v0.0.4
	github.com/olivere/elastic v6.2.21+incompatible
	github.com/olivere/elastic/v7 v7.0.21
	github.com/opentracing/opentracing-go v1.2.0
	github.com/otiai10/copy v1.1.1
	github.com/pborman/uuid v0.0.0-20180906182336-adf5a7427709
	github.com/pierrec/lz4 v0.0.0-20190701081048-057d66e894a4 // indirect
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	github.com/uber-go/tally v3.3.15+incompatible
	github.com/uber/cadence-idl v0.0.0-20211125000611-a6d3e3393ba6
	github.com/uber/ringpop-go v0.8.5
	github.com/uber/tchannel-go v1.22.0
	github.com/uber/tcheck v1.1.0
	github.com/urfave/cli v1.22.4
	github.com/valyala/fastjson v1.4.1
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2
	go.mongodb.org/mongo-driver v1.7.3
	go.opencensus.io v0.22.5 // indirect
	go.uber.org/atomic v1.7.0
	go.uber.org/cadence v0.19.0
	go.uber.org/config v1.4.0
	go.uber.org/fx v1.13.1
	go.uber.org/multierr v1.6.0
	go.uber.org/thriftrw v1.29.2
	go.uber.org/yarpc v1.58.0
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.1.5
	gonum.org/v1/gonum v0.7.0
	google.golang.org/api v0.26.0
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201201144952-b05cb90ed32e // indirect
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.3.0 // indirect
	gopkg.in/validator.v2 v2.0.0-20180514200540-135c24b11c19
	gopkg.in/yaml.v2 v2.2.8
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)

// ringpop-go and tchannel-go depends on older version of thrift, yarpc brings up newer version
replace github.com/apache/thrift => github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7
