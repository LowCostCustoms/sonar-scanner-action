#!/bin/bash -ex

image_name="sonar-scanner-$(uudigen)"
trap "docker image rm $image_name:latest || true"

echo "::group::Build docker image"
docker build -a BASE_IMAGE=$IMAGE -t $image_name .
echo "::endgroup::"

echo "::group::Run action"
docker run \
    --rm \
    -e SONAR_HOST_URL \
    -e SONAR_HOST_CERTIFICATE \
    -e PROJECT_FILE_LOCATION \
    -e WAIT_FOR_QUALITY_GATE \
    -e QUALITY_GATE_WAIT_TIMEOUT \
    -w "$SOURCES_MOUNT_POINT" \
    -v "$SOURCES_LOCATION:$SOURCES_MOUNT_POINT" \
    $image_name
echo "::endgroup::"
