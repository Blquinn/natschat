#!/usr/bin/env bash

migrate -source file://resources/db/migrations -database postgres://postgres:password@localhost:5432/chat?sslmode=disable up
