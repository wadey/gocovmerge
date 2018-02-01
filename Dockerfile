FROM scratch
MAINTAINER Wade Simmons <wade@wades.im> (@wadey)
ADD build/linux/gocovmerge /usr/bin/gocovmerge
ENTRYPOINT ["gocovmerge"]

# vim:syn=Dockerfile:ft=Dockerfile
