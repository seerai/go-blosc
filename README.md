go-blosc
========
[![Build](https://github.com/seerai/go-blosc/actions/workflows/build-test.yaml/badge.svg)](https://github.com/seerai/go-blosc/actions/workflows/build-test.yaml)
[![codecov](https://codecov.io/gh/seerai/go-blosc/branch/main/graph/badge.svg?token=OC1YAKRR0Q)](https://codecov.io/gh/seerai/go-blosc)

golang (cgo) wrapper for [blosc](http://blosc.org/) *"a high performance, multi-threaded, blocking and shuffling
compressor."*

This supports very very basic compression functionality as needed internally for SeerAI's Zarr support. 

This package requires that the blosc library (and headers, if building) are installed. 

On Ubuntu: `apt-get install libblosc-dev`
On Alpine: `apk add blosc blosc-dev`