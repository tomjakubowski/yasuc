# yasuc

yet another [sprunge.us](http://sprunge.us/) clone

(command line pastebin)

## server usage

``` bash
$ go get github.com/tomjakubowski/yasuc
$ yasuc -db /path/to/db -port PORT -addr BIND-ADDR
```

A Dockerfile is provided.

``` bash
$ cd $GOPATH/github.com/tomjakubowski/yasuc
$ sudo docker build -t yasuc/yasuc:v0 .
$ sudo docker run -d -P --name yasuc-test yasuc/yasuc:v0
```

The `docker run` command will map a randomized port on the host to each of the
container's exposed ports (just one right now).  You can find the port using
`docker port`:

``` bash
$ sudo docker port yasuc-test
8080/tcp -> 0.0.0.0:32768
$ curl http://localhost:32768/
usage message goes here
```

## client usage

``` bash
<command> | curl -F 'sprunge=<-' http://my.yasuc.host/
```

## examples

``` bash
$ echo 'hello world' | curl -F 'sprunge=<-' http://my.yasuc.host/
http://my.yasuc.host/a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447
$ curl http://my.yasuc.host/a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447
hello world
```

Each paste is identified by the SHA-256 digest of its contents.

## copyright

Copyright (c) Tom Jakubowski.  License AGPLv3: Affero GPL Version 3.  This is
free software; you are free to change and redistribute it.  There is NO
WARRANTY, to the extent permitted by law.  For copying conditions and source,
see the LICENSE.txt file.
