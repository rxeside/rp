FROM gcr.io/distroless/static-debian12
WORKDIR /app
ADD bin/product /app/product
COPY data /app/data
ENTRYPOINT ["/app/product"]