#############################################
# Build
#############################################
FROM --platform=$BUILDPLATFORM golang:1.19-alpine as build

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git curl

WORKDIR /go/src/github.com/webdevops/helm-azure-tpl

# Dependencies
COPY go.mod go.sum .
RUN go mod download

COPY . .
#RUN make test

# Compile
ARG TARGETARCH
RUN GOARCH=${TARGETARCH} make build
RUN chmod +x entrypoint.sh

#############################################
# Test
#############################################
FROM gcr.io/distroless/static as test
USER 0:0
WORKDIR /app
COPY --from=build /go/src/github.com/webdevops/helm-azure-tpl/helm-azure-tpl .
COPY --from=build /go/src/github.com/webdevops/helm-azure-tpl/entrypoint.sh .
RUN ["./helm-azure-tpl", "--help"]

#############################################
# Final
#############################################
FROM mcr.microsoft.com/azure-cli:latest
ENV LOG_JSON=1
WORKDIR /
COPY --from=test /app .
USER 1000:1000
ENTRYPOINT ["/entrypoint.sh"]
