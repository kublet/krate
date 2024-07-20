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

To update the `krate` binary, run `brew update && brew install kublet/tools/krate`.

### Windows/Linux

Download the `krate` binary from the latest [release](https://github.com/kublet/krate/releases/latest).

Alternatively, you can build from [source](https://github.com/kublet/krate).

If you installed `krate` via brew, `krate` should be available system wide, and you can simply run the `krate` command from anywhere.

If you're on windows, please make the `krate` command available system wide: see [docs](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-7.4#set-environment-variables-in-the-system-control-panel).

## Usage

### Build

```bash
$ krate build
```

### Flash

Run this within the project folder.

```bash
$ krate send <ip-address>
```
