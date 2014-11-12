FROM debian:jessie
RUN apt-get update && apt-get install -y ca-certificates
COPY docker-grid /usr/local/bin/docker-grid
ENTRYPOINT ["/usr/local/bin/docker-grid"]
EXPOSE 8080
CMD ["-h"]
