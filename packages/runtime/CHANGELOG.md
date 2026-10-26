# Changelog

All notable changes to GlassOps Runtime will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

The runtime is a Go rewrite of an earlier TypeScript action that has been removed. No version has been released or tagged, so there are no release notes yet.

Current state:

- Go action in `packages/runtime`, run as a Docker container action
- Policy load, freeze window check and optional code analyzer step
- JWT authentication to Salesforce
- Permit and contract JSON files written to `.glassops/`
