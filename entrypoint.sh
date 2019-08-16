#!/bin/bash

export POD_NAME=`kubectl get pods --selector=service=frontend -o jsonpath='{.items[0].metadata.name}'`
wavefront