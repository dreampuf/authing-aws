FROM golang:latest
ENV GOPATH="/go"
WORKDIR /go/src/github.com/dreampuf/authing-aws
ADD go.mod go.sum ./
RUN go mod tidy
COPY / .
RUN go build -ldflags=-s -o authing-aws

FROM chromedp/headless-shell:latest
RUN apt-get update -y \
    && apt-get install -y ca-certificates bash git dumb-init \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
COPY --from=0 /go/src/github.com/dreampuf/authing-aws/authing-aws /bin/
ENTRYPOINT ["dumb-init", "--"]
RUN mkdir /opt
WORKDIR /opt
