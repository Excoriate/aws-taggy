{
  description = "AWS Taggy Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        # Development shell configuration
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_23

            # Development and build tools
            goreleaser
            golangci-lint
            just

            # Additional utilities
            bash
            git

            # Optional: AWS CLI if needed
            awscli2
          ];

          # Environment variables and shell configuration
          shellHook = ''
            echo "🚀 Welcome to AWS Taggy Development Environment 🏷️"
            echo "Go version: $(go version)"
            echo "Golangci-lint version: $(golangci-lint version)"
          '';
        };

        # Optional: Add build and package configurations
        packages.default = pkgs.buildGoModule {
          pname = "aws-taggy";
          version = "0.1.0";

          src = ./.;

          vendorHash = null;  # Update if needed

          buildPhase = ''
            go build -o $out/bin/aws-taggy
          '';
        };
      }
    );
}
