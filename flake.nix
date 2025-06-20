{
  description = "BurnDevice - 设备破坏性测试应用";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            protobuf
            protoc-gen-go
            protoc-gen-go-grpc
            buf
            grpcurl
            curl
            jq
            git
            gnumake
            neovim
            # Security scanning tools
            gosec
            govulncheck
          ];

          shellHook = ''
            echo "🔥 BurnDevice 开发环境已加载"
            echo "⚠️  警告：此工具仅用于授权测试环境"
            echo ""
            echo "可用工具："
            echo "  - Go $(go version | cut -d' ' -f3)"
            echo "  - protoc $(protoc --version)"
            echo "  - buf $(buf --version)"
            echo "  - gosec $(gosec --version 2>/dev/null | head -1 || echo 'v2.x')"
            echo "  - govulncheck $(govulncheck -version 2>/dev/null || echo 'latest')"
            echo ""
            export GO111MODULE=on
            export GOPROXY=https://goproxy.cn,direct
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "burndevice";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
        };
      });
} 