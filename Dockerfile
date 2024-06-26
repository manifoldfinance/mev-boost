# syntax=docker/dockerfile:1
FROM golang:1.22 as builder
ARG VERSION
WORKDIR /build

COPY go.mod ./
COPY go.sum ./

RUN go mod download

ADD . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -v \
    -ldflags "-w -s -X 'github.com/flashbots/mev-boost/config.Version=$VERSION'" \
    -o mev-boost .

FROM alpine:3.15
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/mev-boost /app/mev-boost
EXPOSE 18550
ENTRYPOINT ["/app/mev-boost"]

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="MEV Boost" \
      org.label-schema.description="MEV Boost Alpine" \
      org.label-schema.url="https://github.com/manifoldfinance/mev-boost" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/manifoldfinance/mev-boost.git" \
      org.label-schema.vendor="CommodityStream, Inc." \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"
