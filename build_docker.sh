#!/bin/bash

VER=0.2.0
IMAGEID="crustio/karst:$VER"
docker build -t $IMAGEID .
