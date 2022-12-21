#!/usr/bin/sh

./geronimo init && \
./geronimo new-account -n "morzona" -u "234234" -r "erg3rg34g" && \
./geronimo new-account -n "mate" -u "53214352" -r "fsgbsfber" && \
./geronimo new-account -n "eszter" -u "34526egs" -r "gfdfsgbasdfr" && \
./geronimo new-broker -n "buffet" -p "ADA/USD" -b 1000 -q 500 -s active -a morzona && \
./geronimo new-broker -n "warren" -b 2000 -q 1500 -s disabled -a mate && \
./geronimo new-broker -n "gekko" -b 100 -q 50 -s active -a mate && \
./geronimo new-broker -n "mobius" -b 200 -q 150 -s active -a eszter && \
./geronimo update-broker -n mobius -s disabled