# Go Deploy (WIP) 
[![Test](https://github.com/mehdibo/go_deploy/actions/workflows/test.yml/badge.svg?branch=develop)](https://github.com/mehdibo/go_deploy/actions/workflows/test.yml)

Go Deploy is a small tool that allows you to listen on webhooks and run SSH commands
or send HTTP requests, it is meant to be used with CI tools to automatically deploy new software versions.

It is ideal to use in an infrastructure that has many internal servers, by using Go Deploy
you will only have to expose one server to listen on webhook requests.
