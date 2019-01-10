# Docker common things

#
# INPUT VARIABLES
# 	- QUAY_USER: The quay.io user to use (usually set in CI)
# 	- QUAY_PASSWD: The quay passwd to use  (usually set in CI)
# 	- IMAGE: the docker image to use. will be computed if it doesn't exist.
# 	- REGISTRY: The docker registry to use. defaults to quay.
#
# EXPORT VARIABLES
# 	- BUILD_NUM: The build number for this build. Will use pants default sandbox
# 	             if not on circleCI, if that isn't available will defauilt to 'dev'.
# 	             If it is in circle will use CIRCLE_BUILD_NUM otherwise.
# 	- IMAGE: The image to use for the build.
# 	- REGISTRY: The registry to use for the build.
# 	- IMAGE_BASENAME: The image without the tag field on it.. i.e. foo:1.0.0 would have image basename of 'foo'
#
#-------------------------------------------------------------------------------
# use pants if it exists outside of circle to get the default namespace and use it for the build
ifndef CIRCLECI
  # @TODO replace/merge this with the newer strategy for providing this in common.mk
  ## create unique build tag
  BUILD_NUM := $(shell pants config get default-sandbox-name 2> /dev/null || echo dev)-$(shell git rev-parse --short HEAD)
endif
ifndef BUILD_NUM
  BUILD_NUM := dev
endif

# TODO: the docker login -e email flag logic can be removed when all projects stop using circleci 1.0 or
#       if circleci 1.0 build container upgrades its docker > 1.14
ifdef CIRCLE_BUILD_NUM
  BUILD_NUM := $(CIRCLE_BUILD_NUM)
  ifeq (email-required, $(shell docker login --help | grep -q Email && echo email-required))
    QUAY := docker login -p "$$QUAY_PASSWD" -u "$$QUAY_USER" -e "unused@unused" quay.io
  else
    QUAY := docker login -p "$$QUAY_PASSWD" -u "$$QUAY_USER" quay.io
  endif
endif

# These can be overridden
REGISTRY ?= quay.io/getpantheon
IMAGE		 ?= $(REGISTRY)/$(APP):$(BUILD_NUM)
# Should we try to pull instead of building?
DOCKER_TRY_PULL ?= false
# Should we rebuild the tag regardless of whether it exists locally or elsewhere?
DOCKER_FORCE_BUILD ?= true

# because users can supply image, we substring extract the image base name
IMAGE_BASENAME := $(firstword $(subst :, ,$(IMAGE)))

# if there is a docker file then set the docker variable so things can trigger off it
ifneq ("$(wildcard Dockerfile))","")
# file is there
  DOCKER:=true
endif

build-docker:: setup-quay build-linux ## build the docker container
	@FORCE_BUILD=$(DOCKER_FORCE_BUILD) TRY_PULL=$(DOCKER_TRY_PULL) ./devops/make/sh/build-docker.sh $(IMAGE)

# stub build-linux std target
build-linux::

push:: setup-quay ## push the container to the registry
	$(call INFO,"pushing image $(IMAGE)")
	@docker push $(IMAGE) > /dev/null

setup-quay:: ## setup docker login for quay.io
  ifdef CIRCLE_BUILD_NUM
    ifndef QUAY_PASSWD
			$(call ERROR, "Need to set QUAY_PASSWD environment variable.")
    endif
  ifndef QUAY_USER
		$(call ERROR, "Need to set QUAY_USER environment variable.")
  endif
	$(call INFO, "Setting up quay login credentials.")
	@$(QUAY) > /dev/null
else
	$(call INFO, "No docker login unless we are in CI.")
	$(call INFO, "We will fail if the docker config.json does not have the quay credentials.")
endif

# we call make here to ensure new states are detected
push-circle:: ## build and push the container from circle
	$(call INFO, "building container before pushing")
	@make build-docker
push-circle:: setup-quay
	@make push

.PHONY:: setup-quay build-docker push push-circle
