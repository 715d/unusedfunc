# Nix Installation and Usage

This document explains how to install and use `unusedfunc` with Nix. The local source flake builds successfully; the NUR package remains an installation option.

## Installation Methods

### Via NUR

`unusedfunc` is available through [NUR](https://github.com/nix-community/NUR) as `nur.repos.715d.unusedfunc`. The current NUR expression packages the v0.2.0 release archive for Linux and Darwin on x86_64 and aarch64.

#### NixOS Configuration

Add to your `configuration.nix`:

```nix
{ pkgs, ... }:
{
  nixpkgs.config.packageOverrides = pkgs: {
    nur = import (builtins.fetchTarball "https://github.com/nix-community/NUR/archive/main.tar.gz") {
      inherit pkgs;
    };
  };

  environment.systemPackages = with pkgs; [
    nur.repos.715d.unusedfunc
  ];
}
```

For a reproducible configuration, use a pinned `fetchTarball` revision and hash instead of the moving `main` archive.

#### Home Manager

Add to your Home Manager configuration:

```nix
{ pkgs, ... }:
{
  nixpkgs.config.packageOverrides = pkgs: {
    nur = import (builtins.fetchTarball "https://github.com/nix-community/NUR/archive/main.tar.gz") {
      inherit pkgs;
    };
  };

  home.packages = with pkgs; [
    nur.repos.715d.unusedfunc
  ];
}
```

#### Command-Line Installation

After configuring NUR in `~/.config/nixpkgs/config.nix` or your system configuration, install to your user profile:

```bash
nix-env -iA nur.repos.715d.unusedfunc -f '<nixpkgs>'
```

### Via Nix Flakes

The repository flake exposes `packages.<system>.default`, `packages.<system>.unusedfunc`, `apps.<system>.default`, and `devShells.<system>.default` for `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, and `aarch64-darwin`.

#### Source-Build Status

The checked-out tree builds successfully with `nix build`. The flake pins a nixpkgs revision that provides Go 1.27 and uses that toolchain for both the package and development shell. Its `vendorHash` matches the module set from this checkout.

This verification does not evaluate the remote `github:715d/unusedfunc` input. A remote ref can have different `go.mod`, `flake.lock`, or `vendorHash` content. Evaluate the target ref before depending on it.

#### Using in Development Environment

Add to your `flake.nix`:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    unusedfunc.url = "github:715d/unusedfunc";
  };

  outputs = { self, nixpkgs, unusedfunc }:
    let
      system = "x86_64-linux";  # or your system
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShells.${system}.default = pkgs.mkShell {
        packages = [
          unusedfunc.packages.${system}.default
        ];
      };
    };
}
```

The remote flake source can differ from this checkout. Evaluate its target ref before using it in production.

## Building from Source

### Build Locally

Clone the repository and build:

```bash
git clone https://github.com/715d/unusedfunc.git
cd unusedfunc
nix build
```

The binary is available at `./result/bin/unusedfunc`.

### Run Tests

```bash
nix develop
make test
```

The development shell provides Go 1.27, the same toolchain used by the package derivation.

## Development

### Development Shell

Enter a development environment with Go:

```bash
nix develop
```

This provides:
- Go compiler and toolchain

Once in the shell, use the Makefile as normal. Development tools are declared in `go.mod` and run through `go tool`; no separate tool installation is needed:

```bash
make build    # Build the binary
make test     # Run tests
make lint     # Run linters
```

### Build Configuration

The Nix build uses these linker values:
- **Version**: `self.rev`, or `dev` when unavailable
- **Git Commit**: `self.shortRev`, or `unknown` when unavailable
- **Build Time**: Fixed to `1970-01-01_00:00:00` for reproducibility
- **ldflags**: `-w -s` for smaller binaries

The Makefile declares matching metadata variables, but its `build` target does not pass `LDFLAGS` to `go build`; binaries made by `make build` retain the command's defaults (`dev`, `unknown`, and `unknown`). The flake sets `doCheck = false`, so a successful `nix build` would not run Go tests.

## Supported Systems

The flake supports the following systems:
- `x86_64-linux` (Linux on Intel/AMD 64-bit)
- `aarch64-linux` (Linux on ARM 64-bit)
- `x86_64-darwin` (macOS on Intel)
- `aarch64-darwin` (macOS on Apple Silicon)

## CI Integration

### GitHub Actions

Use the published NUR flake package in a CI pipeline:

```yaml
name: Lint
on: [push, pull_request]

jobs:
  unusedfunc:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cachix/install-nix-action@v31
      - run: nix profile install github:715d/nur#unusedfunc
      - run: unusedfunc ./...
```

The release configuration publishes to the `715d/nur` repository. Its package currently fetches a release archive, rather than using the repository flake's source build.

## Reproducibility

Nix builds isolate the build environment from system state. In this repository:
- `flake.lock` pins flake inputs
- `go.mod`, `go.sum`, and the package `vendorHash` determine Go dependency inputs
- Build time is fixed for deterministic linker metadata

To update flake inputs:

```bash
nix flake update
```

## Troubleshooting

### Build Fails with "dirty Git tree" Warning

This warning identifies uncommitted changes. It does not by itself indicate a build failure.

### vendorHash Mismatch

If you modify `go.mod` or dependencies, update the `vendorHash`:

1. Set `vendorHash = pkgs.lib.fakeHash;` in `flake.nix`
2. Run `nix build`
3. Copy the correct hash from the error message
4. Update `vendorHash` in `flake.nix`

### Cannot Find Binary After Install

Ensure your Nix profile is in `$PATH`:

```bash
export PATH="$HOME/.nix-profile/bin:$PATH"
```

For permanent setup, add to your shell config (`.bashrc`, `.zshrc`, etc.).

## Comparison with Other Installation Methods

| Method | Pros | Cons |
|--------|------|------|
| **NUR** | Packages the current published release | Requires NUR setup; release version may lag source |
| **Nix Flakes** | Isolated, multi-system source definition | Must evaluate each remote ref before use |
| **go install** | Simple, fast | No Nix profile integration |
| **Binary Release** | No Go toolchain required | Platform-specific |

## Advanced Usage

### Pin a Specific Version

The NUR package currently references the v0.2.0 release archive. Verify each source-flake revision independently before use.

### Use in a Project Flake

Add the source flake to a project's `flake.nix`:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    unusedfunc.url = "github:715d/unusedfunc";
  };

  outputs = { nixpkgs, unusedfunc, ... }:
    let
      system = "x86_64-linux"; # or your system
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      devShells.${system}.default = pkgs.mkShell {
        packages = [ unusedfunc.packages.${system}.default ];
      };
    };
}
```

## Contributing

To contribute Nix-related improvements:

1. Evaluate declared outputs: `nix flake show --no-write-lock-file`
2. Refresh `vendorHash` after dependency changes
3. Verify the package build, then update documentation as needed
4. Submit PR with Nix changes

The CI and release workflows read the Go version from `go.mod`. The Nix package and development shell use Go 1.27.

See `flake.nix` for the complete build configuration.
