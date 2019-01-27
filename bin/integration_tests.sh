#!/usr/bin/env bash

export TEST_DB=chat_test_db

set -e

if [[ ! -z "$(psql -lqt | cut -d \| -f 1 | grep ${TEST_DB})" ]]; then
    read -p "DB ${TEST_DB} already exists, re-create? (yes/NO)" ANSWER
    if [[ ${ANSWER} == "yes" ]]; then
        echo "> Continuing..."
        echo "> Dropping old test db"
        dropdb ${TEST_DB}
    else
        echo "> Aborting"
        exit 1
    fi
fi

echo "> Creating test db"
createdb ${TEST_DB}

echo "> Running migrations"
migrate -source file://resources/db/migrations -database postgres://ben:password@localhost:5432/${TEST_DB} up

echo "> Running test"
go test $* -tags integration

echo "> Destroying test db"
dropdb ${TEST_DB}

set +e

echo "> Done"
