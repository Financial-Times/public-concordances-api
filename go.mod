module github.com/Financial-Times/public-concordances-api

go 1.15

require (
	github.com/Financial-Times/concepts-rw-neo4j v0.0.0-20190415124820-a484da9db19a
	github.com/Financial-Times/go-fthealth v0.0.0-20171204124831-1b007e2b37b7
	github.com/Financial-Times/go-logger v0.0.0-20180323124113-febee6537e90
	github.com/Financial-Times/http-handlers-go v0.0.0-20180517120644-2c20324ab887
	github.com/Financial-Times/neo-model-utils-go v0.0.0-20180712095719-aea1e95c8305
	github.com/Financial-Times/neo-utils-go v0.0.0-20180807105745-1fe6ae2f38f3
	github.com/Financial-Times/service-status-go v0.0.0-20160323111542-3f5199736a3d
	github.com/Financial-Times/transactionid-utils-go v0.2.0 // indirect
	github.com/Financial-Times/up-rw-app-api-go v0.0.0-20170710125828-d9d93a1f6895 // indirect
	github.com/bradfitz/slice v0.0.0-20180809154707-2b758aa73013 // indirect
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.1
	github.com/hashicorp/go-version v1.1.0 // indirect
	github.com/jawher/mow.cli v1.0.5
	github.com/jmcvetta/neoism v1.3.2
	github.com/jmcvetta/randutil v0.0.0-20150817122601-2bb1b664bcff // indirect
	github.com/joho/godotenv v1.3.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	github.com/sirupsen/logrus v1.4.1
	github.com/stretchr/testify v1.3.0
	go4.org v0.0.0-20190313082347-94abd6928b1d // indirect
	golang.org/x/net v0.0.0-20190415100556-4a65cf94b679 // indirect
	golang.org/x/sys v0.0.0-20190415081028-16da32be82c5 // indirect
	gopkg.in/jmcvetta/napping.v3 v3.2.0 // indirect
)

replace github.com/jmcvetta/neoism v1.3.2 => github.com/Financial-Times/neoism v1.3.2-0.20180622150314-0a3ba1ab89c4
