SHELL = /bin/bash

NAME = vindaloo

BUILD_BASE = ./build
BUILD_DIR = ${BUILD_BASE}/opt/${NAME}
BIN_DIR = ${BUILD_DIR}/bin
CONF_DIR = ${BUILD_DIR}/etc

DESCRIPTION = Inventory System
VERSION = $(shell cat etc/vindaloo.json.sample | grep version | cut -d ' ' -f 6 | sed "s/\"//g" )
EPOCH = $(shell date +%s)
URL = https://github.com/euforia/vindaloo

.clean:
	rm -rf ${BUILD_BASE}
	go clean -i ./...

.deps:
	go get -v -d ./...
	[ -d "${BIN_DIR}" ] || mkdir -p ${BIN_DIR}

.test:
	go test -cover ./...

.build_darwin: .clean .deps
	GOOS=darwin go build -v -o ${BIN_DIR}/${NAME}.darwin ./

.build_linux: .clean .deps
	GOOS=linux go build -v -o ${BIN_DIR}/${NAME}.linux ./

.build_native: .clean .deps
	go build -v -o ${BIN_DIR}/${NAME} ./

.build:
	cp -v ./scripts/${NAME}-ctl ${BIN_DIR}/

	mkdir ${BUILD_DIR}/log
	echo "${VERSION}-${EPOCH}" > ${BUILD_DIR}/version

	[ -d "${CONF_DIR}" ] || mkdir ${CONF_DIR}
	cp -rv etc/*.sample ${CONF_DIR}/
	cp -a etc/mappings ${CONF_DIR}/
	cp etc/gnatsd.conf ${CONF_DIR}/
	cp etc/htpasswd ${CONF_DIR}/

	[ -d ${BUILD_BASE}/etc/nginx/conf.d ] || mkdir -p ${BUILD_BASE}/etc/nginx/conf.d
	cp etc/nginx/conf.d/${NAME}.conf ${BUILD_BASE}/etc/nginx/conf.d/

	find ./build -name '.DS_Store' -exec rm -vf '{}' \;

.deb:
	find ./build -name '.DS_Store' -exec rm -vf '{}' \; 
	cd ${BUILD_BASE} && \
	fpm -s dir -t deb --log error -n ${NAME} --epoch ${EPOCH} --version ${VERSION} --description "${DESCRIPTION}" \
		--license MIT --url "${URL}" ./etc ./opt

.rpm:
	find ./build -name '.DS_Store' -exec rm -vf '{}' \;
	cd ${BUILD_BASE} && \
	fpm -s dir -t rpm --log error -n ${NAME} --epoch ${EPOCH} --version ${VERSION} --description "${DESCRIPTION}" \
		--license MIT --url "${URL}" ./etc ./opt

.packages: .deb .rpm

all: .build_native .build
