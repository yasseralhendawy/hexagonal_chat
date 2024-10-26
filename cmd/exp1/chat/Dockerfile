FROM golang:alpine3.19 as builder

WORKDIR /app

COPY / ./


# for the chat service 
RUN GOOS=linux go build -o exp1/chat ./cmd/exp1/chat/main.go


FROM alpine as chat-service
COPY --from=builder /app/exp1/chat exp1/chat
COPY --from=builder /app/cmd/exp1/exp1.yml exp1/
CMD [ "/exp1/chat"]

