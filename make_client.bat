@echo off
SET CGO_ENABLED=0
SET GOOS=windows
go build client.go