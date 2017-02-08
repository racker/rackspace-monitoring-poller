
GIT_TAG := $(shell git describe --abbrev=0)
TAG_DISTANCE := $(shell git describe --long | awk -F- '{print $$2}')
SRC_DIR := pkg
DEB_SRC_DIR := ${SRC_DIR}/debian
BUILD_DIR := build
DEB_BUILD_DIR := ${BUILD_DIR}/debian
EXE := rackspace-monitoring-poller
APP_NAME := rackspace-monitoring-poller

PKGDIR_BIN := usr/bin
PKGDIR_ETC := etc

OS := linux
ARCH := amd64
BIN_URL := https://github.com/racker/rackspace-monitoring-poller/releases/download/$(GIT_TAG)/$(EXE)_$(OS)_$(ARCH)

PKG_DEB := ${BUILD_DIR}/${APP_NAME}_${GIT_TAG}-${TAG_DISTANCE}_${ARCH}.deb

OWNED_DIRS :=
CONFIG_FILES := ${PKGDIR_ETC}/${APP_NAME}.cfg

WGET := wget
FPM := fpm

.PHONY: package package-deb prep clean

package: prep package-deb

package-deb: ${PKG_DEB}

${PKG_DEB} : $(addprefix ${DEB_BUILD_DIR}/,${CONFIG_FILES})
	${FPM} -p $@ -s dir -t deb -C ${DEB_BUILD_DIR} \
	  -n ${APP_NAME} \
	  -v ${GIT_TAG} --iteration ${TAG_DISTANCE} \
	  $(foreach d,${OWNED_DIRS},--directories ${d}) \
	  $(foreach c,${CONFIG_FILES},--config-files ${c}) \
	  ${PKGDIR_BIN}/${EXE} ${CONFIG_FILES}

prep: ${BUILD_DIR}

clean:
	rm -rf $(BUILD_DIR)

${DEB_BUILD_DIR}/${PKGDIR_BIN}/${EXE} : ${DEB_BUILD_DIR}/${PKGDIR_BIN}
	$(WGET) -q -O $@ $(BIN_URL)
	chmod +x $@

${DEB_BUILD_DIR}/${PKGDIR_ETC}/${APP_NAME}.cfg : ${DEB_BUILD_DIR}/${PKGDIR_ETC}
	touch $@

${DEB_BUILD_DIR}/${PKGDIR_BIN} ${DEB_BUILD_DIR}/${PKGDIR_ETC} :
	mkdir -p $@

${BUILD_DIR}:
	mkdir -p $@