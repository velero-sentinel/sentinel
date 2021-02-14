FROM golang:1.15-alpine3.13 as BUILD
ARG BUILD_DATE
ARG REPO_URL
ARG GIT_REVISION
ARG VERSION

WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags \ 
  "-X 'cmd.GitCommit=${GIT_REVISION}' -X 'cmd.Version=${VERSION}' -X 'cmd.BuildTime=${BUILD_DATE}'" \
  -a -o sentinel .

FROM alpine:3.13
ARG BUILD_DATE
ARG REPO_URL
ARG GIT_REVISION
ARG VERSION
LABEL org.opencontainers.image.authors="Markus W Mahlberg <markus@mahlberg.io>"
LABEL org.opencontainers.image.vendor="mahlberg.io"
LABEL org.opencontainers.image.ref.name="velero-sentinel/seninel:${VERSION}"
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.licenses=Apache-2.0
LABEL org.opencontainers.image.url=${REPO_URL}
LABEL org.opencontainers.image.documentation=${REPO_URL}
LABEL org.opencontainers.image.revision=${GIT_REVISION}
LABEL org.opencontainers.image.source=${REPO_URL}
COPY --from=build /go/src/app/sentinel /usr/local/bin/sentinel
ENTRYPOINT [ "/usr/local/bin/sentinel" ]
CMD [ "server" ]