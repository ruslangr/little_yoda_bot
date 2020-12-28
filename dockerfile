FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get github.com/go-telegram-bot-api/telegram-bot-api
RUN go get github.com/fatih/structs
RUN go get github.com/mitchellh/mapstructure
RUN go build -o main .
CMD ["/app/main"]