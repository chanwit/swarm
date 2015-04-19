export GOPATH="`cygpath -m $PWD/Godeps/_workspace`;$GOPATH"
CGO_ENABLED=0 go install -v -a -tags netgo -ldflags "-w -X github.com/docker/swarm/version.GITCOMMIT `git rev-parse --short HEAD`-aiyara"
