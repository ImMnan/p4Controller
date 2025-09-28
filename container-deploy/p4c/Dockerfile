FROM --platform=linux/amd64 alpine:latest


WORKDIR /usr/local/bin
COPY p4controller p4 ./

RUN addgroup -g 1337 p4group && \
    adduser -u 1337 -G p4group -s /bin/sh -D p4cuser && \
    chown p4cuser:p4group /usr/local/bin/p4controller /usr/local/bin/p4 && \
    chmod +x /usr/local/bin/p4controller /usr/local/bin/p4

USER p4cuser

CMD ["/usr/local/bin/p4controller"]