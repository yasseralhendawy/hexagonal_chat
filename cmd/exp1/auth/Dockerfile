FROM golang:alpine3.19 as builder

WORKDIR /app

COPY go.work ./
COPY ./cmd ./cmd

# for the chat service 
RUN GOOS=linux go build -o exp1/auth ./cmd/exp1/auth/main.go


FROM alpine as auth-service
COPY --from=builder /app/exp1/auth exp1/auth
CMD [ "/exp1/auth" ]

