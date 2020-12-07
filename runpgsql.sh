#!/bin/bash

docker pull postgres:11
if [ ! "$(docker ps -q -f name=pgsql1)" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=pgsql1)" ]; then
        docker rm pgsql1
    fi
    docker run --name=pgsql1 -p 5432:5432 -v "/opt/databases/postgres":/var/lib/postgresql/data -e POSTGRES_PASSWORD=goerd -e POSTGRES_DB=goerd -d postgres:11
    ss -tulpn | grep 5432
fi
