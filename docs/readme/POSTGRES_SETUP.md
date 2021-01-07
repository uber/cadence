# Setup local Postgres with Docker
This document describes how to install latest Postgre locally with Docker.

>Note: Install the docker on your machine before installing the MySQL.
* Make sure any MySQL containers are terminated and removed
```
docker ps -a
docker kill <container_id> && docker rm <container_id> # remove any Postgres containers.
```
* Fetch docker image 
```
docker pull postgres
```
* Run docker container (note the port mapping so that 5432 is exposed locally)
```
mkdir -p ~/docker/volumes/postgres
docker run --rm --name pg-docker -e POSTGRES_PASSWORD=cadence -d -p 5432:5432 -v ~/docker/volumes/postgres:/var/lib/postgresql/data postgres
```
* Log into the container (when prompted for password use the password gotten from last step).
```
psql -h localhost -U postgres -d cadence
```