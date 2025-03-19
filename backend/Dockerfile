FROM golang:latest

# Install the necessary packages
RUN apt-get update && apt-get install -y \
    libnfc6 libnfc-bin libnfc-examples libnfc-dev \
    mpg321 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
