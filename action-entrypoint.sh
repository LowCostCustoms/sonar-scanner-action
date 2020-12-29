#!/bin/bash -e

image_name="sonar-scanner-$(uuidgen)"
trap "docker image rm $image_name || true" EXIT

echo "::group::Building docker image"
docker build --build-arg BASE_IMAGE=$IMAGE -t $image_name .
echo "::endgroup::"

echo "::group::Runing sonar-scanner"
docker run \
    --rm \
    -t \
    -e SONAR_HOST_URL \
    -e PROJECT_FILE_LOCATION \
    -e WAIT_FOR_QUALITY_GATE \
    -e QUALITY_GATE_WAIT_TIMEOUT \
    -e LOG_LEVEL \
    -w "$SOURCES_MOUNT_POINT" \
    -v "$SOURCES_LOCATION:$SOURCES_MOUNT_POINT" \
    $image_name
echo "::endgroup::"
