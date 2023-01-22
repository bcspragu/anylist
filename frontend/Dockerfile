FROM gcr.io/distroless/nodejs16-debian11
COPY ./build/ /app
COPY package.json /app
WORKDIR /app
ENV PUBLIC_API_BASE_URL=http://anylist:8080
CMD ["/app"]
