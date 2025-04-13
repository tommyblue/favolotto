#!/bin/bash
socat PTY,link=/tmp/ttyTMP1,raw,echo=0 PTY,link=/tmp/ttyTMP0,raw,echo=0