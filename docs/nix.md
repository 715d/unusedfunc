# Nix Installation and Usage

This document explains how to install and use `unusedfunc` with Nix.

## Installation Methods

### Via NUR (Recommended)

The recommended way to install `unusedfunc` for Nix users is through [NUR](https://github.com/nix-community/NUR) (Nix User Repository).

#### NixOS Configuration

Add to your `configuration.nix`:

```nix
{ pkgs, ... }:
{
  nixpkgs.config.packageOverrides = pkgs: {
    nur = import (builtins.fetchTarball "https://github.com/nix-community/NUR/archive/master.tar.gz") {
      inherit pkgs;
    };
  };

  environment.systemPackages = with pkgs; [
    nur.repos.715d.unusedfunc
  ];
}
```

#### Home Manager

Add to your Home Manager configuration:

```nix
{ pkgs, ... }:
{
  nixpkgs.config.packageOverrides = pkgs: {
    nur = import (builtins.fetchTarball "https://github.com/nix-community/NUR/archive/master.tar.gz") {
      inherit pkgs;
    };
  };

  home.packages = with pkgs; [
    nur.repos.715d.unusedfunc
  ];
}
```

#### Command-Line Installation

Install to your user profile:

```bash
nix-env -iA nur.repos.715d.unusedfunc -f '<nixpkgs>'
```

### Via Nix Flakes

If you prefer using flakes directly:

#### Quick Run (No Installation)

```bash
nix run github:715d/unusedfunc -- ./...
nix run github:715d/unusedfunc -- -v ./internal
```

#### Install to Profile

```bash
nix profile install github:715d/unusedfunc
```

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

## Building from Source

### Build Locally

Clone the repository and build:

```bash
git clone https://github.com/715d/unusedfunc.git
cd unusedfunc
nix build
```

The binary will be available at `./result/bin/unusedfunc`.

### Run Tests

```bash
nix build
nix develop
make test
```

## Development

### Development Shell

Enter a development environment with Go and all necessary tools:

```bash
nix develop
```

This provides:
- Go compiler and toolchain
- All project dependencies

Once in the shell, use the Makefile as normal:

```bash
make build    # Build the binary
make test     # Run tests
make lint     # Run linters
```

### Build Configuration

The Nix build matches the Makefile configuration:
- **Version**: Extracted from git tags (or "dev" for untagged commits)
- **Git Commit**: Short commit hash from the flake
- **Build Time**: Fixed to `1970-01-01_00:00:00` for reproducibility
- **ldflags**: `-w -s` for smaller binaries

## Supported Systems

The flake supports the following systems:
- `x86_64-linux` (Linux on Intel/AMD 64-bit)
- `aarch64-linux` (Linux on ARM 64-bit)
- `x86_64-darwin` (macOS on Intel)
- `aarch64-darwin` (macOS on Apple Silicon)

## CI Integration

### GitHub Actions

Add `unusedfunc` to your CI pipeline:

```yaml
name: Lint
on: [push, pull_request]

jobs:
  unusedfunc:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: cachix/install-nix-action@v27
        with:
          nix_path: nixpkgs=channel:nixos-unstable

      - name: Run unusedfunc
        run: |
          nix run github:715d/unusedfunc -- ./...
```

### NixOS Configuration

Add to your system configuration:

```nix
# configuration.nix
{ pkgs, ... }:
let
  unusedfunc = pkgs.callPackage (pkgs.fetchFromGitHub {
    owner = "715d";
    repo = "unusedfunc";
    rev = "main";  # or specific tag
    sha256 = "...";
  }) {};
in {
  environment.systemPackages = [ unusedfunc ];
}
```

## Reproducibility

Nix builds are reproducible by design:
- All dependencies are pinned in `flake.lock`
- Build environment is isolated from system state
- Build time is fixed for deterministic builds
- No network access during build phase

To update dependencies:

```bash
nix flake update
```

## Troubleshooting

### Build Fails with "dirty Git tree" Warning

This warning is harmless. It appears when uncommitted changes exist. The build will still succeed.

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
| **Nix Flakes** | Reproducible, multi-system, isolated | Requires Nix setup |
| **go install** | Simple, fast | No version pinning |
| **From Source** | Full control | Manual dependency management |
| **Binary Release** | No dependencies | Platform-specific |

## Advanced Usage

### Pin a Specific Version

Use a specific git tag or commit:

```bash
nix run github:715d/unusedfunc/v1.0.0 -- ./...
nix run github:715d/unusedfunc/abc123 -- ./...
```

### Override Build Parameters

Create a custom build with different settings:

```nix
unusedfunc.override {
  buildGoModule = args: pkgs.buildGoModule (args // {
    doCheck = true;  # Enable tests
  });
}
```

### Use in a Project Flake

Add to your project's `flake.nix`:

```nix
{
  inputs.unusedfunc.url = "github:715d/unusedfunc";

  outputs = { self, nixpkgs, unusedfunc }:
    # Use unusedfunc.packages.${system}.default
}
```

## Contributing

To contribute Nix-related improvements:

1. Test changes locally: `nix build`
2. Verify all systems build: `nix flake check`
3. Update documentation as needed
4. Submit PR with Nix changes

See `flake.nix` for the complete build configuration.
