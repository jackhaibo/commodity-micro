#!/bin/bash

work_dir=$@
ls -l $work_dir

mkdir /usr/java
tar xzf $work_dir/jdk-8u231-linux-x64.tar.gz -C /usr/java/
ls -l /usr/java/
java -version

tar -xvf $work_dir/go1.13.6.linux-amd64.tar.gz -C /opt/ >/dev/null
mkdir -p /opt/goworks/bin
ls -l /opt/
go version

#apt-get update
#apt-get -y install git

mkdir -p /usr/share/jenkins
rm -rf /var/lib/apt/lists/*
cp $work_dir/slave.jar /usr/share/jenkins/slave.jar
cp $work_dir/jenkins-slave /usr/bin/jenkins-slave
chmod +x /usr/bin/jenkins-slave

