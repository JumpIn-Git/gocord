{
  inputs = {
    nixpkgs.url = "flake:nixpkgs";
    syntaqlite= {
      url = "github:lalitmaganti/syntaqlite";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, syntaqlite, ... }: let pkgs = nixpkgs.legacyPackages.x86_64-linux; in {
    devShells.x86_64-linux.default = pkgs.mkShellNoCC {
      packages = [
        pkgs.goose
        pkgs.go
        syntaqlite.packages.x86_64-linux.default
        pkgs.sqlite-interactive
      ];
    };
  };
}
