FROM golang:latest

EXPOSE 9000
COPY main.go test_task/
COPY go.* test_task/
RUN cd test_task && go build -o api

ENTRYPOINT ["test_task/api"]