FROM golang:1.20 AS build-stage

RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o natwin .

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build-stage /app/natwin ./natwin
COPY static/ static/
COPY templates/ templates/

EXPOSE 8080

ENTRYPOINT ["/natwin"]
