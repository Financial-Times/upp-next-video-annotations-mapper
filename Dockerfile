FROM alpine:3.5

COPY . /upp-next-video-annotations-mapper/

RUN apk --no-cache --virtual .build-dependencies add git go libc-dev ca-certificates \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/upp-next-video-annotations-mapper" \
  && cd upp-next-video-annotations-mapper \
  && BUILDINFO_PACKAGE="github.com/Financial-Times/upp-next-video-annotations-mapper/buildinfo." \
  && VERSION="version=$(git describe --tag --always 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="repository=$(git config --get remote.origin.url)" \
  && REVISION="revision=$(git rev-parse HEAD)" \
  && BUILDER="builder=$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && echo $LDFLAGS \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && mv * $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && go get -u github.com/kardianos/govendor \
  && $GOPATH/bin/govendor sync \
  && go get -t ./... \
  && go test -race ./... \
  && go build -ldflags="${LDFLAGS}" \
  && mv upp-next-video-annotations-mapper /upp-next-video-annotations-mapper-app \
  && apk del .build-dependencies \
  && rm -rf $GOPATH /var/cache/apk/*

CMD ["/upp-next-video-annotations-mapper-app"]