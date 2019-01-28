#!/usr/bin/env bash

migrate -source file://resources/db/migrations -database postgres://ben:password@localhost:5432/chat up
