#!/usr/bin/env bash

set -e 
if ! [[ "$0" =~ "scripts/build.sh" ]]; then
    echo "must be run from repository root"
    exit 255
fi

ORG_PATH="reyzar.com/server-api"

build(){
    go build -o "bin/server-api" ${ORG_PATH}/cmd ||return
}

buildRpm() {
    CUR_DIR=`pwd`
    RPMBUILD_DIR=$CUR_DIR/rpmBuild

    touch ~/.rpmmacros
    mv ~/.rpmmacros ~/.rpmmacros.bak
    echo '' > ~/.rpmmacros
    echo "%_topdir $RPMBUILD_DIR" >> ~/.rpmmacros
    echo "%__os_install_post %{nil}" >> ~/.rpmmacros

    mkdir -p $RPMBUILD_DIR/BUILD/server-api
    mkdir -p $RPMBUILD_DIR/SOURCES
    cp -f bin/server-api $RPMBUILD_DIR/BUILD/server-api

    cp -f cmd/config/config.yaml $RPMBUILD_DIR/BUILD/server-api
    cp -f $RPMBUILD_DIR/server-api.service $RPMBUILD_DIR/BUILD/server-api
    tar -C $RPMBUILD_DIR/BUILD/ -zcvf $RPMBUILD_DIR/SOURCES/server-api.tar.gz server-api
    rpmbuild -bb $RPMBUILD_DIR/SPECS/server-api.spec
    mv $RPMBUILD_DIR/RPMS/x86_64/*.rpm .
}

if [ "$1" == "rpm" ]; then
    buildRpm
else
    build
fi