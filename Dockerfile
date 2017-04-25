FROM alpine
RUN apk --update add curl ca-certificates
ENV K8S_VERSION v1.6.3
ENV KUBECONFIG /etc/kubernetes/admin.conf
ENV ARCH amd64
RUN curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/kubeadm > /bin/kubeadm
RUN curl -sSL https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/${ARCH}/kubectl > /bin/kubectl
RUN /bin/chmod +x /bin/kubeadm
RUN /bin/chmod +x /bin/kubectl
ADD kmm /bin/kmm
ENTRYPOINT ["/bin/kmm"]
