#!/bin/bash
set -i

for (( c=1; c<=1000; c++ ))
do
	echo "accounts.authentication.password.failed:1|c" | nc -u -w0 127.0.0.1 8127
done