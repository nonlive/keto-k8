FROM alpine
RUN apk --update add curl ca-certificates bash
ADD k8version.cfg .
ADD bin/docker_downloads.sh /
RUN /docker_downloads.sh
ENV KUBECONFIG /etc/kubernetes/admin.conf
RUN /bin/chmod +x /bin/kubectl
ADD kmm /bin/kmm
ENTRYPOINT ["/bin/kmm"]
