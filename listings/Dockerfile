FROM amazonlinux:2018.03
LABEL maintainer="Michael Hausenblas <hausenbl@amazon.com>"
WORKDIR /app
RUN chown -R 1001:1 /app
USER 1001
COPY listings .
EXPOSE 9898
CMD ["/app/listings"]
