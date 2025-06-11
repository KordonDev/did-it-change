FROM alpine:3.18.4

ADD did-it-change /app/dit-it-change

ENV GIN_MODE=release

# Expose the API port
EXPOSE 8080

CMD ["/app/did-it-change"]
