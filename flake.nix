{
  description =
    "Nikon to USB Webcam. Supports older models that Nikon WU does not. Windows/macOS/Linux. No HDMI capture dongle is needed.";

  inputs = {
    nixpkgs.url =
      # this old rev supports Go 1.14. See also:
      # https://lazamar.co.uk/nix-versions/?channel=nixpkgs-unstable&package=go
      "github:nixos/nixpkgs?rev=136a26be29a9daa04e5f15ee7694e9e92e5a028c";
  };

  outputs = { self, nixpkgs }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux

        # not tested:
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];

      forAllSystems = f:
        nixpkgs.lib.genAttrs allSystems
        (system: f { pkgs = import nixpkgs { inherit system; }; });

      mkProgram = pkgs:
        pkgs.buildGo114Module {
          pname = "mtplvcap";
          nativeBuildInputs = [ pkgs.pkg-config ];
          buildInputs = [ pkgs.libusb1 ];
          version = "1.6.2";
          subPackages = [ ];
          src = ./.;
          vendorSha256 = "sha256-J0YW/86VD9nwQ2bq5op0sRLT5/v7W8WfVJ/mmXU2y6o=";
        };
    in {
      apps = forAllSystems ({ pkgs }: {
        default = {
          type = "app";
          program = "${mkProgram pkgs}/bin/mtplvcap";
        };
      });

      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          name = "mtplvcap-dev-shell";
          buildInputs = with pkgs; [ go libusb1 pkg-config ];
          shellHook = ''
            if [ "''${MTPLVCAP_DEV_SHELL:-x}" == "entered" ]; then
              exit 0
            fi
            echo ""
            echo "This is the dev shell for the MTPLVCAP project."
            echo ""
            export GOROOT="${pkgs.go}/share/go"
            export MTPLVCAP_DEV_SHELL="entered"
          '';
        };
      });

      packages = forAllSystems ({ pkgs }: { default = mkProgram pkgs; });
    };
}
