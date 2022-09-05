# syntax=docker/dockerfile:1
FROM golang:1.18 as builder
ARG VERSION
ARG CGO_CFLAGS
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
WORKDIR /build
ADD . /build/
RUN --mount=type=cache,target=/root/.cache/go-build CGO_CFLAGS="$CGO_CFLAGS" GOOS=linux go build -ldflags "-X 'github.com/flashbots/mev-boost/config.Version=$VERSION'" -v -o mev-boost .

FROM alpine:3.15
RUN apk add --no-cache libstdc++ libc6-comp

RUN apk add --no-cache bind-toolsat
RUN addgroup -g 10001 -S nonroot && adduser -u 10000 -S -G nonroot -h /home/nonroot nonroot

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