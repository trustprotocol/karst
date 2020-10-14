#!/bin/bash

VER=0.4.0
IMAGEID="crustio/karst:$VER"
docker build -t $IMAGEID .
