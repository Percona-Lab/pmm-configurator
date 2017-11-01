#!/usr/bin/env bats

[ -z "$SUT" ] && SUT='http://127.0.0.1:7777' || :
[ -z "$URL_PREFIX" ] && URL_PREFIX='configurator' || :

@test "check update - up-to-date" {
    if [ -n "${REMOTE}" ]; then
        skip "can be checked only locally"
    fi

    echo '# v1.4.0' > ${BATS_TEST_DIRNAME}"/sandbox/main.yml"
    echo '# v1.4.0' > ${BATS_TEST_DIRNAME}"/sandbox/new.yml"

    run curl \
        -s \
        -X GET \
        --insecure \
        "${SUT}/${URL_PREFIX}/v1/check-update"
    echo "$output" >&2
    rm -rf ${BATS_TEST_DIRNAME}"/sandbox/main.yml" ${BATS_TEST_DIRNAME}"/sandbox/new.yml"

    [[ "$output" = '{"code":404,"status":"Not Found","title":"Your PMM version is up-to-date."}' ]]
}

@test "check update - new version available" {
    if [ -n "${REMOTE}" ]; then
        skip "can be checked only locally"
    fi

    echo '# v1.4.0' > ${BATS_TEST_DIRNAME}"/sandbox/main.yml"
    echo '# v1.5.0' > ${BATS_TEST_DIRNAME}"/sandbox/new.yml"

    run curl \
        -s \
        -X GET \
        --insecure \
        "${SUT}/${URL_PREFIX}/v1/check-update"
    echo "$output" >&2
    rm -rf ${BATS_TEST_DIRNAME}"/sandbox/main.yml" ${BATS_TEST_DIRNAME}"/sandbox/new.yml"

    [[ "$output" = '{"code":200,"status":"OK","title":"A new PMM version is available.","from":"1.4.0","to":"1.5.0"}' ]]
}

@test "check update - unknown version available" {
    if [ -n "${REMOTE}" ]; then
        skip "can be checked only locally"
    fi

    echo '# old version' > ${BATS_TEST_DIRNAME}"/sandbox/main.yml"
    echo '# new version' > ${BATS_TEST_DIRNAME}"/sandbox/new.yml"

    run curl \
        -s \
        -X GET \
        --insecure \
        "${SUT}/${URL_PREFIX}/v1/check-update"
    echo "$output" >&2
    rm -rf ${BATS_TEST_DIRNAME}"/sandbox/main.yml" ${BATS_TEST_DIRNAME}"/sandbox/new.yml"

    [[ "$output" = '{"code":200,"status":"OK","title":"A new PMM version is available.","from":"unknown","to":"unknown"}' ]]
}