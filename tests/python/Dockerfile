FROM python:3.12-slim

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

LABEL version="${VERSION}" commit="${COMMIT}" date="${DATE}"

WORKDIR /app

COPY pyproject.toml .
COPY src/ src/

RUN pip install --no-cache-dir .

CMD ["python", "-c", "from src.hello import greet; print(greet('Dagger'))"]
