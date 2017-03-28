FROM alpine
RUN apk --update add curl ca-certificates
ENV K8S_VERSION v1.6.0-rc.1
ENV ARCH amd64
RUN curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/kubeadm > /bin/kubeadm
RUN /bin/chmod +x /bin/kubeadm
ADD kmm /bin/kmm
ENTRYPOINT ["/bin/kmm"]
