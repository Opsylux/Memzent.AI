FROM envoyproxy/envoy:v1.31-latest
COPY envoy.yaml /etc/envoy/envoy.yaml
EXPOSE 50051 9901
CMD ["/usr/local/bin/envoy", "-c", "/etc/envoy/envoy.yaml", "--service-cluster", "memzent-envoy"]
