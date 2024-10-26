FROM golang:alpine3.19 as builder

WORKDIR /app

COPY / ./
RUN GOOS=linux go build -o exp1/auth ./cmd/exp1/auth/main.go


FROM alpine as auth-service
COPY --from=builder /app/exp1/auth exp1/auth
COPY --from=builder /app/cmd/exp1/exp1.yml /
CMD [ "/exp1/auth" ]

