FROM nginx:latest
LABEL MAINTAINER jackhaibo

RUN mkdir /usr/share/nginx/html/resources
COPY resources /usr/share/nginx/html/resources
RUN rm /etc/nginx/conf.d/default.conf && rm /etc/nginx/nginx.conf
ADD nginx.conf /etc/nginx/
RUN /bin/bash -c 'echo init ok!!!'