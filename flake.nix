{
  description = "Sync ACME-issued TLS certificates into Authentik's certificate store";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      overlay = final: prev: {
        authentik-cert-sync = final.buildGoModule {
          pname = "authentik-cert-sync";
          version = "0.1.0";
          src = self;
          vendorHash = null;
        };
      };

      nixosModule = { config, lib, pkgs, ... }:
        let
          cfg = config.services.authentik-cert-sync;
        in
        {
          options.services.authentik-cert-sync = {
            enable = lib.mkEnableOption "authentik-cert-sync";

            authentikUrl = lib.mkOption {
              type = lib.types.str;
              description = "Base URL of the Authentik instance";
            };

            certName = lib.mkOption {
              type = lib.types.str;
              description = "Name of the certificate in Authentik's certificate store";
            };

            acmeDomain = lib.mkOption {
              type = lib.types.str;
              description = "ACME domain whose cert/key will be synced (resolves to /var/lib/acme/<acmeDomain>/{cert,key}.pem)";
            };

            tokenFile = lib.mkOption {
              type = lib.types.str;
              description = "Path to a file containing the Authentik API token";
            };
          };

          config = lib.mkIf cfg.enable {
            systemd.services.authentik-cert-sync = {
              description = "Sync ACME certificate ${cfg.acmeDomain} into Authentik";
              after = [ "acme-finished-${cfg.acmeDomain}.target" "network-online.target" ];
              wants = [ "network-online.target" ];

              serviceConfig = {
                Type = "oneshot";
                ExecStart = lib.escapeShellArgs [
                  "${pkgs.authentik-cert-sync}/bin/authentik-cert-sync"
                  "--url"
                  cfg.authentikUrl
                  "--cert-name"
                  cfg.certName
                  "--cert-file"
                  "/var/lib/acme/${cfg.acmeDomain}/cert.pem"
                  "--key-file"
                  "/var/lib/acme/${cfg.acmeDomain}/key.pem"
                  "--token-file"
                  cfg.tokenFile
                ];
                Restart = "no";
                # Run as the acme user so cert files are readable
                User = "acme";
              };
            };

            # NOTE: Consuming configs must add "authentik-cert-sync" to
            # security.acme.certs.<acmeDomain>.reloadServices so the service
            # triggers on cert renewal. Example:
            #   security.acme.certs.${cfg.acmeDomain}.reloadServices =
            #     [ "authentik-cert-sync.service" ];
          };
        };

    in
    {
      overlays.default = overlay;
      nixosModules.default = nixosModule;
    }
    // flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ overlay ];
        };
      in
      {
        packages.default = pkgs.authentik-cert-sync;

        checks.default = pkgs.authentik-cert-sync.overrideAttrs (_: {
          doCheck = true;
        });

        devShells.default = pkgs.mkShell {
          buildInputs = [ pkgs.go pkgs.gotools ];
        };
      });
}
