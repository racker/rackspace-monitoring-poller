
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
REPREPRO := reprepro

PKGDIR_BIN := usr/bin
PKGDIR_ETC := etc

OS := linux
ARCH := amd64
BIN_URL := https://github.com/racker/rackspace-monitoring-poller/releases/download/$(GIT_TAG)/$(EXE)_$(OS)_$(ARCH)
VENDOR := Rackspace US, Inc.
LICENSE := Apache v2

PKG_DEB := ${BUILD_DIR}/${APP_NAME}_${GIT_TAG}-${TAG_DISTANCE}_${ARCH}.deb

# TODO: should poller get its own specific file?
APP_CFG := ${PKGDIR_ETC}/rackspace-monitoring-agent.cfg
UPSTART_CONF := ${PKGDIR_ETC}/init/${APP_NAME}.conf
UPSTART_DEFAULT := ${PKGDIR_ETC}/default/${APP_NAME}

OWNED_DIRS :=
DEB_CONFIG_FILES := ${APP_CFG}
DEB_ALL_FILES := ${DEB_CONFIG_FILES} ${UPSTART_CONF} ${UPSTART_DEFAULT}

WGET := wget
FPM := fpm

.PHONY: default repackage package package-deb package-repo-upload package-upload-deb package-deb-local clean generate-mocks stage-deb-exe-local

default: clean package

generate-mocks:
	mockgen -source=poller/poller.go -package=poller -destination=poller/poller_mock_test.go
	mockgen -destination mock_golang/mock_conn.go -package mock_golang net Conn

package: package-deb

package-repo-upload: package-deb package-upload-deb

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

${DEB_BUILD_DIR}/${APP_CFG} : ${SRC_DIR}/generic/sample.cfg | ${DEB_BUILD_DIR}/${PKGDIR_ETC}
	cp $< $@

${DEB_BUILD_DIR}/${UPSTART_CONF} : ${DEB_SRC_DIR}/service.upstart ${DEB_BUILD_DIR}/${PKGDIR_ETC}/init
	cp $< $@
	chmod +x $@

${DEB_BUILD_DIR}/${UPSTART_DEFAULT} : ${DEB_SRC_DIR}/upstart_default.cfg | ${DEB_BUILD_DIR}/${PKGDIR_ETC}/default
	cp $< $@

${BUILD_DIR} ${DEB_BUILD_DIR}/${PKGDIR_BIN} ${DEB_BUILD_DIR}/${PKGDIR_ETC} ${DEB_BUILD_DIR}/${PKGDIR_ETC}/init | ${DEB_BUILD_DIR}/${PKGDIR_ETC}/default :
	mkdir -p $@
