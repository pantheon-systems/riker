APP := riker

# compile release binaries for these archs and os's
RELEASE_ARCH := 386 amd64
RELEASE_OS := linux darwin openbsd

include devops/make/common.mk
include devops/make/common-docker.mk
include devops/make/common-docs.mk
include devops/make/common-go.mk

build-release:: ## build binaries for all supported platforms
	@go get github.com/mitchellh/gox
	@rm -rf ./dist/
	@gox \
		-os="$(RELEASE_OS)" \
		-arch="$(RELEASE_ARCH)" \
		-output "./dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

# extend the update-makefiles task to remove files we don't need
update-makefiles::
	make prune-common-make

# strip out everything from common-makefiles that we don't want.
prune-common-make:
	@find devops/make -type f  \
		-not -name common.mk \
		-not -name common-docker.mk \
		-not -name common-docs.mk \
		-not -name common-go.mk \
		-not -name install-go.sh \
		-not -name build-docker.sh \
		-not -name install-gcloud.sh \
		-not -name install-shellcheck.sh \
		-not -name update-kube-object.sh \
		-delete
	@find devops/make -empty -delete
	@git add devops/make
	@git commit -C HEAD --amend
