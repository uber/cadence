# Setup local MySQL with Docker
This document describes how to install MySQL 5.7 locally with Docker.

>Note: Install the docker on your machine before installing the MySQL.
* Make sure any MySQL containers are terminated and removed
```
docker ps -a
docker kill <container_id> && docker rm <container_id> # remove any MySQL containers.
```
* Fetch docker image (version 5.7 is what Travis runs so its what you will want locally)
```
docker pull mysql/mysql-server:5.7
```
* Run docker container (note the port mapping so that 3306 is exposed locally)
```
docker run -p 3306:3306 --name=mysql1 -d mysql/mysql-server:5.7
```
* When docker starts up the root MySQL user will have an auto generated password. You need to get that password to log into the container
```
docker logs mysql1 2>&1 | grep GENERATED
# The result looks like: [Entrypoint] GENERATED ROOT PASSWORD: iHqEvRYm6UP#YN$es;YnV3m(oJ
```
* Log into the container (when prompted for password use the password gotten from last step).
```
docker exec -it mysql1 mysql -uroot -p
```
* Before any SQL operations can be performed you must reset the root user's password (use anything you like in replace of root_password).
```
SET PASSWORD = PASSWORD('root_password');
```
* Now create the user that local MySQL tests will use. Also grant all privileges to user.
```
CREATE USER 'uber'@'%' IDENTIFIED BY 'uber';
GRANT ALL PRIVILEGES ON *.* TO 'uber'@'%';
```