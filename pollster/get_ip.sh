#!/bin/sh

ifconfig $1 | grep "inet addr" | awk '{print $2}' | awk -F: '{print $2}'

