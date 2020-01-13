#!/bin/bash

ls

cd gateway/run && docker build -t gateway:v1.0 .
cd -

cd item/run && docker build -t item:v1.0 .
cd -

cd order/run && docker build -t order:v1.0 .
cd -

cd user/run && docker build -t user:v1.0 .
cd -

cd view && docker build -t view:v1.0 .
cd -

ls