module github.com/Financial-Times/public-concordances-api

go 1.15

require (
	github.com/Financial-Times/api-endpoint v1.0.0
	github.com/Financial-Times/concepts-rw-neo4j v1.26.0
	github.com/Financial-Times/go-fthealth v0.0.0-20171204124831-1b007e2b37b7
	github.com/Financial-Times/go-logger/v2 v2.0.1
	github.com/Financial-Times/http-handlers-go/v2 v2.3.0
	github.com/Financial-Times/neo-model-utils-go v1.0.0
	github.com/Financial-Times/neo-utils-go/v2 v2.0.0
	github.com/Financial-Times/service-status-go v0.0.0-20160323111542-3f5199736a3d
	github.com/Financial-Times/transactionid-utils-go v0.2.0
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.1
	github.com/jawher/mow.cli v1.0.5
	github.com/jmcvetta/neoism v1.3.2
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/sirupsen/logrus v1.4.1 // indirect
	github.com/stretchr/testify v1.6.1
	go4.org v0.0.0-20190313082347-94abd6928b1d // indirect
	golang.org/x/net v0.0.0-20190415100556-4a65cf94b679 // indirect
	golang.org/x/sys v0.0.0-20190415081028-16da32be82c5 // indirect
)

replace github.com/jmcvetta/neoism v1.3.2 => github.com/Financial-Times/neoism v1.3.2-0.20180622150314-0a3ba1ab89c4
