FROM alpine:3.20.0 AS build
RUN apk update && apk add --no-cache make go
WORKDIR /stun/
COPY . /stun
RUN ["make", "build"]

FROM alpine:3.20.0
RUN apk add tini
WORKDIR /stun
COPY --from=build /stun/bin/stun .
COPY docker/start.sh .
RUN ["chmod", "+x", "start.sh"]
ENTRYPOINT ["tini", "--", "./start.sh"]
CMD ["./stun"]
