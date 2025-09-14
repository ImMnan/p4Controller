FROM debian:latest


WORKDIR /usr/local/bin
COPY p4controller.sh p4d p4 ./

RUN groupadd -g 1337 p4group && useradd -u 1337 -g p4group -s /bin/sh -m p4cuser \
    && chown p4cuser:p4group /usr/local/bin/p4controller.sh \
    && chown p4cuser:p4group /usr/local/bin/p4 \
    && chown p4cuser:p4group /usr/local/bin/p4d \
    && chmod +x /usr/local/bin/p4controller.sh /usr/local/bin/p4d /usr/local/bin/p4

RUN mkdir -p /var/p4d-root /var/p4d-checkpoint/ /var/p4d-versionfiles/ \
    && chown -R p4cuser:p4group /var/p4d-root /var/p4d-checkpoint/ /var/p4d-versionfiles/

USER p4cuser

CMD ["/usr/local/bin/p4controller.sh"]