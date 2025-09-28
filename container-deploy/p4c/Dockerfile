FROM debian:bookworm-slim

WORKDIR /usr/local/bin
COPY p4controller p4 ./

RUN addgroup --gid 1337 p4group && \
    adduser --uid 1337 --ingroup p4group --shell /bin/sh --disabled-password p4cuser && \
    chown p4cuser:p4group /usr/local/bin/p4controller /usr/local/bin/p4 && \
    chmod +x /usr/local/bin/p4controller /usr/local/bin/p4

USER p4cuser

CMD ["/usr/local/bin/p4controller"]