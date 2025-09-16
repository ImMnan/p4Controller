FROM --platform=linux/amd64 debian:latest

ENV ROOT_DIR="/opt/p4d-root"
ENV CHECKPOINT_DIR="/opt/p4d-check/checkpoint"
ENV VERSION_DIR="/opt/p4d-ver/versionfile"
ENV P4D_IP=""
ENV P4D_PORT=4232

WORKDIR /usr/local/bin
COPY p4controller.sh p4d p4 ./

RUN groupadd -g 1337 p4group && \
    useradd -u 1337 -g p4group -s /bin/sh -m p4cuser && \
    chown p4cuser:p4group /usr/local/bin/p4controller.sh /usr/local/bin/p4 /usr/local/bin/p4d && \
    chmod +x /usr/local/bin/p4controller.sh /usr/local/bin/p4d /usr/local/bin/p4 
#    mkdir -p $ROOT_DIR/ $VERSION_DIR/versionfile/ $CHECKPOINT_DIR/checkpoint/ && \
#    chown -R p4cuser:p4group $ROOT_DIR/ $VERSION_DIR/ $CHECKPOINT_DIR/

USER p4cuser

CMD ["/usr/local/bin/p4controller.sh"]