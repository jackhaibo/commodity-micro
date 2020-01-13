FROM 192.168.134.104/library/jenkins-slave-base:v1.0
LABEL maintainer jackhaibo

ENV JAVA_HOME /usr/java/jdk1.8.0_231
ENV CLASSPATH $JAVA_HOME/lib/tools.jar:$JAVA_HOME/lib/dt.jar:$JAVA_HOME/lib
ENV PATH $JAVA_HOME/bin:$PATH

ENV GOROOT /opt/go
ENV GOPATH /opt/goworks
ENV PATH $PATH:$GOROOT/bin:/$GOPATH/bin

ARG BUILD_DIR=/opt/build
RUN mkdir $BUILD_DIR
COPY go1.13.6.linux-amd64.tar.gz jdk-8u231-linux-x64.tar.gz jenkins-slave settings.xml slave.jar build.sh $BUILD_DIR/
RUN /bin/bash $BUILD_DIR/build.sh $BUILD_DIR

ENTRYPOINT ["jenkins-slave"]
