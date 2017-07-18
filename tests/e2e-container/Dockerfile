FROM quay.io/ukhomeofficedigital/systemd

RUN dnf upgrade -y && \
    dnf install \
    dnsmasq \
    jq \
    which \
    dnf-plugins-core \
    -y && \
    dnf clean all

# Add Docker in Docker...
RUN dnf config-manager \
        --add-repo \
        https://download.docker.com/linux/fedora/docker-ce.repo -y && \
        dnf makecache fast -y && \
        dnf install \
        docker-ce \
        -y  && \
        dnf clean all

ENV container docker
RUN curl -s https://pkg.cfssl.org/R1.1/cfssl_linux-amd64 -o /usr/bin/cfssl && \
    chmod +x /usr/bin/cfssl
RUN curl -s https://pkg.cfssl.org/R1.1/cfssljson_linux-amd64 -o /usr/bin/cfssljson && \
    chmod +x /usr/bin/cfssljson

# We need a specific version to work with what's shipping in CoreOS
# TODO: make the same as CoreOS stable (blocked by bug in 1.25 https://github.com/coreos/bugs/issues/1892)
#       Unfortunatly CoreOS doesn't include the shipping version of the kubelet-wrapper or rkt in release feed
ENV RKT_VER=1.27.0
ENV RKT_RPM=rkt-${RKT_VER}-1.x86_64.rpm
RUN curl -sSL https://github.com/rkt/rkt/releases/download/v${RKT_VER}/${RKT_RPM} > /tmp/${RKT_RPM} && \
    rpm -i /tmp/${RKT_RPM} && \
    rm /tmp/${RKT_RPM}
RUN mkdir -p /opt/bin /usr/lib/coreos /usr/local/lib/rkt/stage1-images
RUN curl -sSL \
    https://github.com/rkt/rkt/releases/download/vv${RKT_VER}/stage1-fly-v${RKT_VER}-linux-amd64.aci \
    > /usr/local/lib/rkt/stage1-images/stage1-fly.aci
RUN curl -sSL \
    https://raw.githubusercontent.com/coreos/coreos-overlay/master/app-admin/kubelet-wrapper/files/kubelet-wrapper \
    > /usr/lib/coreos/kubelet-wrapper && \
    chmod +x /usr/lib/coreos/kubelet-wrapper

VOLUME /var/lib/docker
VOLUME /var/lib/kubelet

CMD ["/usr/sbin/init"]
