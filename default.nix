{ pkgs ? import (builtins.fetchTarball {
    url = "https://github.com/nixos/nixpkgs/archive/729e7295cf7b.tar.gz";
    sha256 = "0mxhi0lc11aa3r7i7din1q2rjg5c4amq3alcr8ga2fcb64vn2zd3";
  }) { } 
}:

pkgs.buildGoModule {
  pname = "go-cast";
  version = "dev";

  src = ./.;

  vendorSha256 = null;
}
