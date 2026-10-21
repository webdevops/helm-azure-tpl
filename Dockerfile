#############################################
# Build
#############################################
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git curl upx

WORKDIR /go/src/github.com/webdevops/helm-azure-tpl

COPY . .
RUN make test

# Compile
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build
RUN upx /go/src/github.com/webdevops/helm-azure-tpl/helm-azure-tpl
RUN chmod +x entrypoint.sh

#############################################
# Test
#############################################
FROM gcr.io/distroless/static AS test
USER 0:0
WORKDIR /app
COPY --from=build /go/src/github.com/webdevops/helm-azure-tpl/helm-azure-tpl .
COPY --from=build /go/src/github.com/webdevops/helm-azure-tpl/entrypoint.sh .
RUN ["./helm-azure-tpl", "--help"]

#############################################
# final
#############################################
FROM gcr.io/distroless/static AS final-static
WORKDIR /
COPY --from=test /app .
USER 1000:1000

#############################################
# final
#############################################
FROM mcr.microsoft.com/azure-cli:latest AS final-azcli
WORKDIR /
COPY --from=test /app .
USER 1000:1000
ENTRYPOINT ["/entrypoint.sh"]
