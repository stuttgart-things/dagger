# Happy-path fixture: passes hadolint with the repo's ignore set.
#
# The curl pin is required -- DL3018 is not in that ignore set -- and it is
# tied to the base image: each alpine tag ships exactly one curl version, and
# `apk add curl=<other>` fails outright. Renovate bumps the FROM tag but cannot
# know about the coupling, so when the tag moves, look up the new version with
# `docker run --rm alpine:<tag> apk policy curl` and update this line too.
FROM alpine:3.24

RUN apk add --no-cache curl=8.22.0-r0

WORKDIR /app

COPY . .

CMD ["sh"]
