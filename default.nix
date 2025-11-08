# NUR compatibility layer
# This allows non-flake users to build unusedfunc
{ pkgs ? import <nixpkgs> { } }:

let
  flake = builtins.getFlake (toString ./.);
in
flake.packages.${pkgs.system}.default
