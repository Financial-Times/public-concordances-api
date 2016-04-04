FROM alpine:3.3
ADD *.go .git /public-concordances-api/
ADD concordances/*.go /public-concordances-api/concordances/
RUN apk add --update bash \
  && apk --update add git go \
  && cd public-concordances-api \
  && git fetch origin 'refs/tags/*:refs/tags/*' \
  && BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo." \
  && VERSION="$(git describe --tag --always 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="$(git config --get remote.origin.url)" \
  && REVISION="$(git rev-parse HEAD)" \
  && BUILDER="$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && cd .. \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/public-concordances-api" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && cp -r public-concordances-api/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && go get ./... \
  && cd $GOPATH/src/${REPO_PATH} \
  && echo ${LDFLAGS} \
  && go build -ldflags="${LDFLAGS}"
  && mv public-concordances-api /app \
  && apk del go git \
  && rm -rf $GOPATH /var/cache/apk/*
CMD exec /app --neo-url=$NEO_URL --port=$APP_PORT --graphiteTCPAddress=$GRAPHITE_ADDRESS --graphitePrefix=$GRAPHITE_PREFIX --logMetrics=$LOG_METRICS --cache-duration=$CACHE_DURATION
