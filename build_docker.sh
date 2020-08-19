#!/bin/bash

VER=0.3.0
IMAGEID="crustio/karst:$VER"
docker build -t $IMAGEID .
