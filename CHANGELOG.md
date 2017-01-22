# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## [1.1.0] - 2017-01-22
### Added
- LDAPv3 user driver
- Method agnostic authentication webservice to work with [NGINX auth_request
  directive](http://nginx.org/en/docs/http/ngx_http_auth_request_module.html#auth_request)
- Service registration and service discovery with etcd

## [1.0.0] - 2017-01-16
### Changed
- Use HTTP instead of GRPC for server-to-sever communication.
