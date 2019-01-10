#! /bin/bash
# install and configure gcloud on circle-ci
#
# The following ENV vars must be set before calling this script:
#
#   GCLOUD_EMAIL           # user-id for circle to authenticate to google cloud
#   GCLOUD_KEY             # base64 encoded key
set -e

if [[ "$CIRCLECI" != "true" ]]; then
  echo "This script is only intended to run on Circle-CI."
  exit 1
fi

export CLOUDSDK_CORE_DISABLE_PROMPTS=1
export CLOUDSDK_PYTHON_SITEPACKAGES=0

# ensure we use certs to talk to kube, instead of the oauth bridge (google auth creds)
export CLOUDSDK_CONTAINER_USE_CLIENT_CERTIFICATE=True

gcloud="$HOME/google-cloud-sdk/bin/gcloud -q --no-user-output-enabled"
PATH="$gcloud/bin:$PATH"

# circle may have an old gcloud installed we wipe it out cause we bring our own.
# Make sure we remove the default install bashrc modifications, otherwise install.sh
# will create an invalid bashrc
if [ -d "/opt/google-cloud-sdk" ] ; then
  sed -ie '/The next line updates PATH/,+3d' "$HOME/.bashrc"
  sed -ie '/The next line enables/,+3d' "$HOME/.bashrc"
  sudo rm -rf /opt/google-cloud-sdk
fi
if [ ! -d "$HOME/google-cloud-sdk" ]; then
  echo "$HOME/google-cloud-sdk missing, installing"
  curl -o "$HOME/google-cloud-sdk.tar.gz" https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.tar.gz
  tar -C "$HOME/" -xzf ~/google-cloud-sdk.tar.gz
  # somehow, .bashrc.backup is owned by root sometimes. This makes `install.sh` fail, so remove it here
  sudo rm -f "$HOME/.bashrc.backup"
  bash "$HOME/google-cloud-sdk/install.sh" --rc-path "$HOME/.bashrc" --quiet

  $gcloud components update
  $gcloud components update kubectl
fi
