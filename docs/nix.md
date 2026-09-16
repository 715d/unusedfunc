# Nix

The checked-in flake provides `packages.default`, `packages.unusedfunc`, `apps.default`, and `devShells.default` for `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, and `aarch64-darwin`.

## Use the Flake

```bash
nix build .#unusedfunc
./result/bin/unusedfunc ./...

nix run .# -- ./...
nix develop
```

The development shell provides Go 1.27. `nix build` does not run tests because the package sets `doCheck = false`; run `make test` in the shell.

`default.nix` supports `nix-build` invocation, but it uses `builtins.getFlake` and still requires flakes:

```bash
nix-build default.nix
```

## Update Dependencies

After you change Go dependencies, set `vendorHash` in `flake.nix` to `pkgs.lib.fakeHash`, run `nix build`, then replace it with the hash from the error. Update flake inputs with:

```bash
nix flake update
```
