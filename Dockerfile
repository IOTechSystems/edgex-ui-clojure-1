FROM alpine:3.8 as gobuider

WORKDIR /root/go

RUN apk add --no-cache go git build-base

COPY src/go . 

RUN go get -d -v github.com/russolsen/transit
RUN go get -d -v github.com/gin-gonic/gin
RUN go get -d -v gopkg.in/resty.v1
RUN go get -d -v github.com/BurntSushi/toml
RUN CGO_ENABLED=0 go install -v -ldflags '-extldflags "-static"' github.com/edgexfoundry/go-ui-server

FROM clojure:lein-2.7.1-alpine as clojurebuilder
COPY . /usr/src/app
WORKDIR /usr/src/app
RUN lein with-profile production cljsbuild once production

FROM scratch
WORKDIR /root/
COPY --from=gobuider /root/go/bin/go-ui-server .
COPY --from=clojurebuilder /usr/src/app/resources/public assets
COPY --from=clojurebuilder /usr/src/app/resources/configuration.toml res/configuration.toml
ENV PORT=8080
EXPOSE $PORT
ENTRYPOINT ["/root/go-ui-server"]
