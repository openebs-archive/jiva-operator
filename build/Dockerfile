FROM @BASEIMAGE@

ARG ARCH

ENV OPERATOR=/usr/local/bin/jiva-operator \
    USER_UID=1001 \
    USER_NAME=jiva-operator

# install operator binary
COPY build/_output/bin/${ARCH}/jiva-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
