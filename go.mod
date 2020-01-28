module github.com/uber/cadence

go 1.12

require (
	cloud.google.com/go/bigquery v1.4.0 // indirect
	cloud.google.com/go/storage v1.5.0
	github.com/Shopify/sarama v1.26.0
	github.com/apache/thrift v0.13.0
	github.com/aws/aws-sdk-go v1.28.9
	github.com/benbjohnson/clock v1.0.0 // indirect
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/cactus/go-statsd-client/statsd v0.0.0-20191106001114-12b4e2b38748
	github.com/cch123/elasticsql v0.0.0-20190321073543-a1a440758eb9
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dgryski/go-farm v0.0.0-20191112170834-c2139c5d712b
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/emirpasic/gods v0.0.0-20190624094223-e689965507ab
	github.com/fatih/color v1.9.0
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocql/gocql v0.0.0-20200121121104-95d072f1b5bb
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/golang/mock v1.4.0
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/jonboulle/clockwork v0.1.0
	github.com/klauspost/compress v1.9.8 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lib/pq v1.3.0
	github.com/m3db/prometheus_client_golang v0.8.1
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mattn/go-sqlite3 v2.0.2+incompatible // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/olivere/elastic v6.2.27+incompatible
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pborman/uuid v1.2.0
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.4.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/robfig/cron v1.2.0
	github.com/samuel/go-thrift v0.0.0-20191111193933-5165175b40af // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/uber-common/bark v1.2.1 // indirect
	github.com/uber-go/kafka-client v0.2.3-0.20191018205945-8b3555b395f9
	github.com/uber-go/tally v3.3.13+incompatible
	github.com/uber/ringpop-go v0.8.5
	github.com/uber/tchannel-go v1.16.0
	github.com/urfave/cli v1.22.2
	github.com/valyala/fastjson v1.4.5
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2
	go.uber.org/atomic v1.5.1
	go.uber.org/cadence v0.9.1-0.20200128004345-b282629d5ba9
	go.uber.org/fx v1.10.0 // indirect
	go.uber.org/goleak v1.0.0 // indirect
	go.uber.org/multierr v1.4.0
	go.uber.org/net/metrics v1.3.0 // indirect
	go.uber.org/thriftrw v1.22.0
	go.uber.org/yarpc v1.42.1
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20200128002243-345141a36859
	google.golang.org/api v0.15.0
	google.golang.org/genproto v0.0.0-20200127141224-2548664c049f // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.4.0 // indirect
	gopkg.in/validator.v2 v2.0.0-20191107172027-c3144fdedc21
	gopkg.in/yaml.v2 v2.2.8
)

// TODO https://github.com/uber/cadence/issues/2863
replace github.com/jmoiron/sqlx v1.2.0 => github.com/longquanzheng/sqlx v0.0.0-20191125235044-053e6130695c
