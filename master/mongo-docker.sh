#!/bin/bash
docker run --name mongo1 -v ./mongo/m1:/etc/mongo -d mongo --config /etc/mongo/mongod.conf
docker run --name mongo2 -v ./mongo/m2:/etc/mongo -d mongo --config /etc/mongo/mongod.conf
docker run --name mongo3 -v ./mongo/m3:/etc/mongo -d mongo --config /etc/mongo/mongod.conf
