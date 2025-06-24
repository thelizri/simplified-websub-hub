FROM golang:1.22

RUN go install github.com/codeskyblue/fswatch@latest

RUN mkdir -p /go/src/hub
WORKDIR /go/src/hub

# Copy fswatch config
COPY fsw.yml /fsw.yml

CMD fswatch -config /fsw.yml