# Sigstore signing

> **Opt-in, and not in a published release yet.** This is planned for v1.1.0. HMAC signing stays the default. See [what is in which release](releases-and-licenses.md#what-is-in-which-release).

By default a PERMIT is signed with a shared secret, `GLASSOPS_PERMIT_KEY`. With Sigstore signing, `gate` signs the PERMIT keyless on GitHub Actions instead. There is no secret to store or rotate. The signature says which workflow, in which repository, signed the PERMIT, and `permit verify` and `deploy` accept only a signature from the workflow and repository you name. The PERMIT itself is unchanged. See [PERMIT and rules](permit-and-rules.md) for what it holds.

GlassOps does not sign anything itself. It runs the separate program `cosign`, which is in both GlassOps images (see [Releases, images and licenses](releases-and-licenses.md#tools-in-the-images)).

---

## Turn it on

```yaml
policy:
  permit:
    signing: sigstore
    identity: https://github.com/acme/sf-app/.github/workflows/deploy.yml@refs/heads/main
    issuer: https://token.actions.githubusercontent.com
    repository: acme/sf-app
```

| Field | Meaning |
|---|---|
| `signing` | `hmac` (the default) or `sigstore` |
| `identity` | The certificate identity a signature must carry. Required for `sigstore`. For GitHub Actions it is `https://github.com/<owner>/<repo>/.github/workflows/<file>@<full ref>`, such as `@refs/heads/main` or `@refs/tags/v1` |
| `issuer` | The OIDC issuer that vouches for the identity. Required for `sigstore`. For GitHub Actions it is `https://token.actions.githubusercontent.com` |
| `repository` | The repository, as `owner/name`, that a signature must come from. On GitHub Actions it defaults to the running repository (`GITHUB_REPOSITORY`). Off GitHub Actions it is required |

`glassops doctor` and every command that loads the config refuse `sigstore` without both `identity` and `issuer`, and a `repository` that is not `owner/name`.

---

## What each command needs

**`gate`** signs. It stops with exit 2 and writes no PERMIT unless all of these hold, checked in this order:

1. The run is on GitHub Actions. Bitbucket Pipelines and GitLab CI are not supported for Sigstore signing. Use HMAC there.
2. The job has an OIDC token. Give the job that runs `gate` the permission `id-token: write`. GlassOps checks only that the token variables exist and never reads their values.
3. `cosign` is on `PATH`.
4. A repository to pin is known.

`gate` does not read `GLASSOPS_PERMIT_KEY` when signing is `sigstore`.

**`permit verify` and `deploy`** verify. They run `cosign verify-blob` and accept the PERMIT only when the signature, the transparency log entry and a certificate for the configured `identity` from the configured `issuer`, issued to a workflow run in the pinned `repository`, all check out. Verifying needs no OIDC token, so it works in a job without `id-token: write` and on a laptop that has `cosign`. Off GitHub Actions you must set `policy.permit.repository`. Verifying is refused on Bitbucket Pipelines and GitLab CI. `cosign` can refresh the Sigstore trust root, so verifying can need network access.

A missing `cosign` or a missing repository is a tool error (exit 2). A signature that does not verify is a failed check in `permit verify` (exit 1) and a refusal (exit 2) in `deploy`, the same as for HMAC. A signature that is not valid base64 or not a cosign bundle is refused before `cosign` runs. Everything else the PERMIT binds is checked as before: target, commit, plan, config, change set, expiry, lifetime and the clean tree rule.

The `signature` block of a Sigstore PERMIT has the algorithm `sigstore-bundle` and no `key_id`. Its value is the cosign bundle in standard base64. Two bundle shapes are accepted. The one that `cosign` v2 writes with `sign-blob --bundle` holds `base64Signature`, `cert` and `rekorBundle`. The newer protobuf bundle carries a `mediaType` that starts with `application/vnd.dev.sigstore.bundle`. Anything else is refused. The images carry `cosign` v2.5.3.

---

## Identity and the repository pin

The certificate identity names the workflow that ran the signing, and it includes the ref that workflow ran at. Two things follow.

**Your own workflow.** When `gate` runs in a workflow file in your repository, the identity is that file at that ref, as in the example above. A deploy from `main` and a deploy from a tag have different identities. Name the one you want to accept, and protect that branch or tag, because anyone who can change the workflow on it can sign.

**A reusable workflow.** When the signing runs inside a workflow your repository calls, the certificate names the called workflow, not yours. Every repository that calls `glassops-deploy.yml@<ref>` would share one identity, `https://github.com/glassops-platform/glassops/.github/workflows/glassops-deploy.yml@<full ref>`. That is why `repository` exists: the identity alone cannot tell one consumer from another, so verifying also passes `--certificate-github-workflow-repository` and accepts only a certificate issued to a run in your repository. Do not leave `repository` out when you set `identity` to a reusable workflow, and do not rely on `identity` alone.

The reusable workflow `glassops-deploy.yml` grants `id-token: write` to its deploy job and passes the OIDC request variables into the container that runs `gate`, so a call to it can sign with Sigstore. The caller must grant `id-token: write` on the job that calls the workflow, because a called workflow cannot hold more than its caller grants. A caller that grants less is refused when the run starts, whichever signing it uses. With HMAC signing the permission goes unused and nothing else changes. `glassops doctor` recognises a call to the reusable workflow and tells you which identity such a call would carry. You can also run `gate` in a workflow of your own that grants the permission.

---

## What goes into the public log

Keyless signing writes to the public Rekor transparency log. Each signing records the SHA-256 of the PERMIT and the signing certificate, never the PERMIT itself. The certificate holds the repository name, the workflow ref and the commit. **If your repository is private, its name, the workflow path and the commit hash become public.** Anyone can read the log. Use HMAC signing if that is not acceptable.

---

## Doctor checks

With `policy.permit.signing: sigstore`, `glassops doctor` adds these checks. It adds none for HMAC.

| Check | What it says |
|---|---|
| `tool.cosign` | Fails when `cosign` is not found. Otherwise it reports the version |
| `permit.sigstore` | Fails on Bitbucket Pipelines or GitLab CI. On GitHub Actions it passes when the job can request an OIDC token and fails when it cannot. A pull request job without the token gets a warning only. Off any CI it is skipped, because only verifying works there |
| `permit.identity` | Passes with a `warning:` detail when the issuer is not the GitHub one, when the identity is not shaped like a GitHub workflow identity, or when your workflows call `glassops-deploy.yml` at a ref the identity does not name |
| `permit.repository` | Reports the repository that will be pinned and where it came from. It warns when none is known, or when the configured one differs from the running repository |

---

## Tested here, and not

- The signer and verifier are covered by unit tests with a fake `cosign`, including a wrong identity, a wrong issuer, a wrong repository, a tampered PERMIT, a missing OIDC token, a missing `cosign` and a non-GitHub CI. Sigstore PERMITs carry and check the same change set and command hashes as HMAC ones.
- One real signing run on GitHub Actions with `cosign` v2.5.3 and a dummy file confirmed that `cosign` finds the job's OIDC token itself, that `--bundle` writes the older bundle shape, and that the certificate identity is the called workflow and not the caller. That run wrote one public Rekor entry for the dummy file.
- That a Sigstore signature is produced and verified end to end through a call to the reusable `glassops-deploy.yml`, and that the pinned repository there is the caller's repository, comes from reading the workflow and `cosign`'s options. It has not been proven on a real run yet.
