#!/bin/bash -e

docker compose run --remove-orphans --build --rm lint
