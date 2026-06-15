# Regression fixture: trips hadolint rules NOT in the repo ignore set.
# - DL3006: tag the base image (here it's untagged -> :latest)
# - DL3008: pin versions in apt-get install
# - DL3009: delete the apt-get lists after installing
FROM ubuntu

RUN apt-get update && apt-get install -y curl

CMD ["bash"]
