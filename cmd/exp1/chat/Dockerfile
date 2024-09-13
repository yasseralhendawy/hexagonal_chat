FROM golang:alpine3.19 as builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY ./cmd ./cmd
COPY ./config ./config

# for the chat service 
RUN GOOS=linux go build -o exp1/chat ./cmd/exp1/chat/main.go


FROM alpine as chat-service
COPY --from=builder /app/exp1/chat exp1/chat
CMD [ "/exp1/chat" ]

