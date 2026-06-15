# Happy-path fixture: passes hadolint with the repo's ignore set.
FROM alpine:3.21

RUN apk add --no-cache curl=8.12.1-r1

WORKDIR /app

COPY . .

CMD ["sh"]
