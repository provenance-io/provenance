# Registry class examples

This directory holds example registry classes that can be created with the registry CLI.

## `example_registry_class.json`

A complete example registry class expressing the participant role policies as ordinary
`role_authorizations` data: Originator, Lien Owner (with foreclosure),
Controller (with foreclosure), Secured Party for Lien, Secured Party for eNote, Pledgee, Servicer,
and Sub-servicer.

These policies are evaluated by the registry's authorization engine; none of them are hard-coded into
the chain. A maintainer installs them per asset class:

```bash
provenanced tx registry create-registry-class example_registry_class.json --from <maintainer-key>
```

Before submitting, set `registry_class_id`, `asset_class_id`, and `maintainer` to your values (the
`maintainer` must equal the `--from` signer, or it can be omitted to default to that signer).

The file is kept in sync with the policy builders in
`x/registry/keeper/participant_role_policies_acceptance_test.go`; regenerate it with:

```bash
REGEN_EXAMPLE_FIXTURE=1 go test ./x/registry/keeper/ -run TestParticipantRolePoliciesAcceptanceTestSuite/TestExampleFixtureInSync
```
