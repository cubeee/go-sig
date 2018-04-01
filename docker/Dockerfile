FROM golang:latest as builder
WORKDIR /go/src/github.com/cubeee/go-sig
COPY ./install_dependencies.sh $PWD
COPY ./dependencies.txt $PWD
RUN $PWD/install_dependencies.sh
COPY ./ $PWD
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o build/go-sig .

FROM alpine:latest
RUN apk --no-cache add ca-certificates bash
WORKDIR /root/
COPY --from=builder /go/src/github.com/cubeee/go-sig/build/go-sig ./
COPY --from=builder /go/src/github.com/cubeee/go-sig/resources ./resources/
RUN chmod +x /root/go-sig;
CMD ["/root/go-sig"]