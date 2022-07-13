# Swagger

This readme contains information related to the Swagger UI for Provenance.

<!-- TOC -->
  - [Activation](#activation)
  - [Location](#location)
  - [Updating](#updating)



## Activation

In order to have the Swagger UI available on your node, it must be turned on in the config.

1. If the file `build/run/provenanced/config/app.toml` does not yet exist run this command:
    ```bash
    > make build run-config
    ```
2. Enable the api and swagger features:
    ```bash
    > build/provenanced --home build/run/provenanced config set api.enable true api.swagger true
    ```
3. Start the node:
    ```bash
    > make run
    ```
4. If you are refreshing the browser, hard refresh, e.g. `cmd` + `shift` + `r`.



## Location

The swagger UI can be found at the `/swagger/` path on your node.

For example, if running locally, it will be at http://localhost:1317/swagger/

If you don't get any response (or you get a connection refused):
1. Make sure it has been [activated](#activation):
    ```bash
    > build/provenanced --home build/run/provenanced config get api.enabled api.swagger
    ```
2. Make sure your firewall is allowing traffic port `1317`.
3. Make sure that port `1317` is the port defined in the `api.address` config value:
    ```bash
    > build/provenanced --home build/run/provenanced config get api.address
    ```



## Updating

Any time changes are made to the proto files (`/proto` or `third_party/proto`), the Swagger UI files need to be updated too.

If the Cosmos-SDK version has been changed since the last time the swagger files were generated, the `swagger_third_party.yaml` file needs to be updated first.

To get a new version of `swagger_third_party.yaml`:

1. In this repo, identify the version of Cosmos-SDK being used.
2. Clone [our fork of the Cosmos-SDK](https://github.com/provenance-io/cosmos-sdk) repo locally and navigate to it.
3. In that repo, check out the tag for the version identified earlier.
4. In that repo, run `make proto-swagger-gen`.
5. Copy the resulting `client/docs/swagger-ui/swagger.yaml` file from that repo to `client/docs/swagger_third_party.yaml` in this repo.

Finally, to update the rest of the swagger files (regardless of whether the 3rd party file needed updating), run this command:
```bash
> make update-swagger-docs
```

Commit any files with changes so they can be included in a PR.

You will need to `make build` again in order to see any changes in your browser.
