# Usage Guide

This document provides detailed usage instructions for the Proof CLI.

## Basic Commands

### Help

```bash
./proof --help
```

This displays the help information for the CLI, including available commands and flags.

### Version

```bash
./proof version
```

Displays the current version of the CLI.

## Configuration

The CLI can be configured using a configuration file. By default, it looks for a `.proof-cli.yaml` file in your home directory.

You can specify a custom configuration file using the `--config` flag:

```bash
./proof --config /path/to/config.yaml
```

### Configuration Format

The configuration file uses YAML format:

```yaml
api_endpoint: https://api.example.com
timeout: 30
```

## Environment Variables

All configuration options can also be set using environment variables. The CLI will automatically look for environment variables that match the configuration keys, prefixed with `PROOF_CLI_`.

For example:

```bash
export PROOF_CLI_API_ENDPOINT=https://api.example.com
export PROOF_CLI_TIMEOUT=30
export PROOF_CLI_DEBUG=true
```