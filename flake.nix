{
  inputs = {
    nixpkgs.url = "flake:nixpkgs";
  };

  outputs = { nixpkgs, ... }: let pkgs = nixpkgs.legacyPackages.x86_64-linux; in {
    devShells.x86_64-linux.default = pkgs.mkShellNoCC {
      packages = [
        pkgs.goose
        pkgs.go
      ];
    };
  };
}
