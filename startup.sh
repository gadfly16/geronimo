#!/usr/bin/sh

./geronimo init && \
./geronimo new-account -n "morzona" -k "234234" -p "erg3rg34g" && \
./geronimo new-account -n "mate" -k "53214352" -p "fsgbsfber" && \
./geronimo new-account -n "eszter" -k "34526egs" -p "gfdfsgbasdfr" && \
./geronimo new-broker -n "buffet" -b 1000 -q 500 -s active -a morzona && \
./geronimo new-broker -n "warren" -b 2000 -q 1500 -s disabled -a mate && \
./geronimo new-broker -n "gekko" -b 100 -q 50 -s active -a mate && \
./geronimo new-broker -n "mobius" -b 200 -q 150 -s active -a eszter && \
./geronimo update-broker -n mobius -s disabled