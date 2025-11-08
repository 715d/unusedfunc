{
  description = "Go linter detecting unused functions with /internal awareness";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ] (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Extract version from git tags, fallback to "dev"
        version = if self ? rev then self.rev else "dev";

        # Use short commit hash or "unknown"
        gitCommit = if self ? shortRev then self.shortRev else "unknown";

        # Fixed build time for reproducibility
        buildTime = "1970-01-01_00:00:00";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "unusedfunc";
          inherit version;

          src = ./.;

          vendorHash = "sha256-LhXcU9vyMJliPg9rvx6FY24GiJALGSu2g6vovyOlkXw=";

          # Inject version info matching Makefile pattern
          ldflags = [
            "-X main.version=${version}"
            "-X main.buildTime=${buildTime}"
            "-X main.gitCommit=${gitCommit}"
            "-w"
            "-s"
          ];

          # Tests require network for some testdata
          doCheck = false;

          meta = with pkgs.lib; {
            description = "Go linter detecting unused functions with /internal awareness";
            homepage = "https://github.com/715d/unusedfunc";
            license = licenses.bsd3;
            maintainers = [ ];
            mainProgram = "unusedfunc";
          };
        };

        # Development shell with Go toolchain
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
          ];

          shellHook = ''
            echo "unusedfunc development environment"
            echo "Go version: $(go version)"
          '';
        };

        # Convenience aliases
        packages.unusedfunc = self.packages.${system}.default;
        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/unusedfunc";
        };
      }
    );
}
