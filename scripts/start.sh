#!/bin/sh
go run ./cmd/main.go --web-log-mode=debug --db-type=postgres --db-connection="host=localhost user=iliaf password=iliaf dbname=iliaf port=5432 sslmode=disable"