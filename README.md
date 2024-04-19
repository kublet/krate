# Krate

[![Docs](https://img.shields.io/badge/docs-developers.thekublet-blue?style=flat-square)](https://developers.thekublet.com)

Krate is a CLI development tool for building and flashing firmware onto [Kublet devices](https://thekublet.com).

## Requirements

- WiFi
- Visual Studio Code
- PlatformIO (PIO)
- Kublet

`Krate` must be used within a PIO terminal.

## Install

### MacOS

```bash
$ brew tap kublet/tools
$ brew install krate 
```

### Windows/Linux

Download the `krate` binary from the latest [release](https://github.com/kublet/krate/releases/latest).

Alternatively, you can build from [source](https://github.com/kublet/krate).

## Usage

### Build

```bash
$ krate build
```

### Flash

```bash
$ krate send <ip-address>
```
