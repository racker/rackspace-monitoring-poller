
GIT_TAG := $(shell git describe --abbrev=0)
TAG_DISTANCE := $(shell git describe --long | awk -F- '{print $$2}')
SRC_DIR := pkg
DEB_SRC_DIR := ${SRC_DIR}/debian
DEB_REPO_DIR := ${DEB_SRC_DIR}/repo
CLOUDFILES_REPO_NAME := poller-$(GIT_TAG)
BUILD_DIR := build
DEB_BUILD_DIR := ${BUILD_DIR}/debian
EXE := rackspace-monitoring-poller
APP_NAME := rackspace-monitoring-poller
PROJECT_VENDOR := github.com/racker/rackspace-monitoring-poller/vendor

PKGDIR_BIN := usr/bin
PKGDIR_ETC := etc

OS := linux
ARCH := amd64
BIN_URL := https://github.com/racker/rackspace-monitoring-poller/releases/download/$(GIT_TAG)/$(EXE)_$(OS)_$(ARCH)
VENDOR := Rackspace US, Inc.
LICENSE := Apache v2

PKG_DEB := ${BUILD_DIR}/${APP_NAME}_${GIT_TAG}-${TAG_DISTANCE}_${ARCH}.deb

# TODO: should poller get its own specific file?
APP_CFG := ${PKGDIR_ETC}/rackspace-monitoring-poller.cfg
UPSTART_CONF := ${PKGDIR_ETC}/init/${APP_NAME}.conf
UPSTART_DEFAULT := ${PKGDIR_ETC}/default/${APP_NAME}
LOGROTATE_CFG := ${PKGDIR_ETC}/logrotate.d/${APP_NAME}

OWNED_DIRS :=
DEB_CONFIG_FILES := ${APP_CFG} ${LOGROTATE_CFG}
DEB_ALL_FILES := ${DEB_CONFIG_FILES} ${UPSTART_CONF} ${UPSTART_DEFAULT}

MOCK_POLLER := LogPrefixGetter,ConnectionStream,Connection,Session,CheckScheduler,CheckExecutor,Scheduler,ChecksReconciler

WGET := wget
FPM := fpm
REPREPRO := reprepro

.PHONY: default repackage package package-deb package-repo-upload package-upload-deb package-deb-local \
	clean generate-mocks stage-deb-exe-local prep \
	generate-callgraphs regenerate-callgraphs clean-callgraphs

default: clean package

generate-mocks:
	mockgen -package=poller_test -destination=poller/poller_mock_test.go github.com/racker/rackspace-monitoring-poller/poller ${MOCK_POLLER}
	mockgen -source=utils/events.go -package=utils -destination=utils/events_mock_test.go
	mockgen -destination check/pinger_mock_test.go -package=check github.com/racker/rackspace-monitoring-poller/check Pinger
	sed -i '' s,$(PROJECT_VENDOR)/,, check/pinger_mock_test.go
	mockgen -destination mock_golang/mock_conn.go -package mock_golang net Conn

prep:
	curl https://glide.sh/get | sh
	${GOPATH}/bin/glide install

regenerate-callgraphs : clean-callgraphs generate-callgraphs

generate-callgraphs : docs/poller_callgraph.png docs/endpoint_callgraph.png

clean-callgraphs :
	rm -f docs/*_callgraph.{dot,png}

%_callgraph.png : %_callgraph.dot
	dot -Tpng -o $@ $<

%_callgraph.dot : ${GOPATH}/bin/go-callvis
	${GOPATH}/bin/go-callvis -focus $(*F) -nostd -group type github.com/racker/rackspace-monitoring-poller > $@

${GOPATH}/bin/go-callvis :
	go get -u github.com/TrueFurby/go-callvis

package: package-deb

package-repo-upload: package-deb reprepro-deb package-upload-deb

package-upload-deb:
	rclone mkdir rackspace:${CLOUDFILES_REPO_NAME}/debian
	rclone copy ${DEB_REPO_DIR}/ rackspace:${CLOUDFILES_REPO_NAME}/debian

reprepro-deb:
	${REPREPRO} -b ${DEB_REPO_DIR} includedeb cloudmonitoring build/*.deb

package-deb: ${PKG_DEB}

package-deb-local: stage-deb-exe-local package-deb

${PKG_DEB} : ${DEB_BUILD_DIR}/${PKGDIR_BIN}/${EXE} $(addprefix ${DEB_BUILD_DIR}/,${DEB_ALL_FILES}) ${DEB_BUILD_DIR}
	rm -f $@
	${FPM} -p $@ -s dir -t deb \
	  -n ${APP_NAME} --license "${LICENSE}" --vendor "${VENDOR}" \
	  -v ${GIT_TAG} --iteration ${TAG_DISTANCE} \
	  $(foreach d,${OWNED_DIRS},--directories ${d}) \
	  $(foreach c,${DEB_CONFIG_FILES},--config-files ${c}) \
	  --deb-default ${DEB_BUILD_DIR}/${UPSTART_DEFAULT} \
	  --deb-upstart ${DEB_BUILD_DIR}/${UPSTART_CONF} \
	  -C ${DEB_BUILD_DIR} ${PKGDIR_BIN}/${EXE} ${DEB_CONFIG_FILES}

clean:
	rm -rf $(BUILD_DIR)

stage-deb-exe-local: | ${DEB_BUILD_DIR}/${PKGDIR_BIN}
	cp -p ${BUILD_DIR}/${APP_NAME}_${OS}_${ARCH} ${DEB_BUILD_DIR}/${PKGDIR_BIN}/${EXE}

${DEB_BUILD_DIR}/${PKGDIR_BIN}/${EXE} : | ${DEB_BUILD_DIR}/${PKGDIR_BIN}
	$(WGET) -q --no-use-server-timestamps -O $@ $(BIN_URL)
	chmod +x $@

${DEB_BUILD_DIR}/${APP_CFG} : ${SRC_DIR}/generic/sample.cfg ${DEB_BUILD_DIR}/${PKGDIR_ETC}
	cp $< $@

${DEB_BUILD_DIR}/${LOGROTATE_CFG} : ${SRC_DIR}/generic/logrotate.cfg ${DEB_BUILD_DIR}/${PKGDIR_ETC}/logrotate.d
	cp $< $@

${DEB_BUILD_DIR}/${UPSTART_CONF} : ${DEB_SRC_DIR}/service.upstart ${DEB_BUILD_DIR}/${PKGDIR_ETC}/init
	cp $< $@
	chmod +x $@

${DEB_BUILD_DIR}/${UPSTART_DEFAULT} : ${DEB_SRC_DIR}/upstart_default.cfg ${DEB_BUILD_DIR}/${PKGDIR_ETC}/default
	cp $< $@

${BUILD_DIR} $(addprefix ${DEB_BUILD_DIR}/,${PKGDIR_BIN} ${PKGDIR_ETC} ${PKGDIR_ETC}/logrotate.d ${PKGDIR_ETC}/init ${PKGDIR_ETC}/default) :
	mkdir -p $@
