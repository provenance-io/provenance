# Swagger

This readme contains information related to the Swagger UI for Provenance.

## Activation

In order to have the Swagger UI available on your node, it must be turned on in the config.

1.  If the file `build/run/provenanced/config/app.toml` does not yet exist run this command:
    ```bash
    > make build run-config
    ``` 
1.  Open the `build/run/provenanced/config/app.toml` for editing.
    1.  Find the `API Configuration` section.
    1.  In that section, change `enable = false` to `enable = true`.
    1.  Also in that section, change `swagger = false` to `swagger = true`.
1. Regenerate the swagger files:
    ```bash
    > make update-swagger-docs
    ```
1. Rebuild provenance: 
   ```bash
   > make build
   ```
1. Start the node:
    ```bash
    > make run
    ```
1. If you are refreshing the browser, hard refresh:
    `cmd` + `shift` + `r`

## Location

The swagger UI can be found at the `/swagger/` path on your node.

For example, if running locally, it will be at http://localhost:1317/swagger/

If you don't get any response (or you get a connection refused):
1.  Make sure it has been [activated](#activation).
1.  Make sure your firewall is allowing traffic port `1317`.
1.  Make sure that port `1317` is still the correct port; it's defined in the `app.toml` file's `[api]` section in the `address` field.


## Updating

Any time changes are made to the proto files, the Swagger UI files need to be updated too.

To update the swagger docs:
```bash
> make update-swagger-docs
```

This will update a couple files.
Those changes should be committed and included in your PR.
