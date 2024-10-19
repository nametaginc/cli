# Command-line tools for Nametag

# Install

`nametag` is a command-line utility that lets you work with Nametag using commands, from configuring accounts and environments to performing directory integrations. It runs on your local device so you'll want to install the version that's appropriate for your operating system.

## MacOS

If you have the [Homebrew](https://brew.sh) package manager installed, `nametag` can be installed by running:

```
$ brew install nametaginc/tap/nametag
```

If not, you can run the install script:

```
$ curl -L https://nametag.co/install.sh | sh
```

If you used curl to install `nametag`, then you need to add the nametag directory to your shell rc file. Check the output of the install script for the entries to copy and paste into the file. Now you can use the `nametag` command from any directory.

## Linux

Run the install script:

```
$ curl -L https://nametag.co/install.sh | sh
```

## Windows

Run the PowerShell install script:

```
$ pwsh -Command "iwr https://nametag.co/install.ps1 -useb | iex"
```

If you encounter an error saying the `pwsh` command is not found, `powershell` can be used instead, though we recommend [installing the latest version of PowerShell](https://learn.microsoft.com/en-us/powershell/scripting/install/installing-powershell-on-windows).

# Configure Authentication

You can use a Nametag API key to authenticate by setting the environment variable `NAMETAG_AUTH_TOKEN` of the `--auth-token` command line argument.

# Next Steps

Run `nametag --help` for a tour of the available commands
