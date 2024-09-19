FROM golang:1.18
WORKDIR /usr/src/app
COPY ./main.go ./
RUN go mod init app
RUN go mod tidy
ENTRYPOINT ["go", "run", "main.go"]
