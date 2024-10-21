#!/bin/bash -e

docker compose build test --no-cache
docker compose run test

