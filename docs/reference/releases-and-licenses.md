# Releases, images and licenses

This page describes how a GlassOps release is cut, which image tags exist, and how the license scan and the third party notices work.

## What is in which release

v0.3.0 and v0.4.0 are published prereleases on the [releases page](https://github.com/glassops-platform/glassops/releases). v1.0.0 is not out yet. This table says which release first has each command or feature. "Not yet" means it is in the repository and in no published release, so you need a build from this repository to use it. See [Install](cli.md#install).

| Command or feature | First release |
|---|---|
| `init`, `doctor`, `describe`, `plan`, `validate` | v0.3.0 |
| `gate`, `permit verify`, `deploy`, `report` | v0.3.0 |
| `env create`, `env delete`, `env sweep` | v0.4.0 |
| `test`, `check`, `analyze` and the `glassops-analyzer` image | v0.4.0 |
| `run` and the default `pr` and `deploy` pipelines | v0.4.0 |
| The reusable GitHub workflows and the consumer example | v0.4.0 |
| `verify`, and `verify` in the default `deploy` pipeline | Not yet |
| Provider `exec` and the engine result schema | Not yet |
| The Bitbucket and GitLab CI layers | Not yet |
| `env create` applying definition `settings` | Not yet |
| `test.coverage.class_scope` | Not yet |
| The hardened PERMIT: `change_set_sha256`, the lifetime and key checks, `gate --deploy-command` | Not yet |
| `deploy` refusing a tree that is not the PERMIT commit, the window cap on `expires_at`, and `--resume` limits | Not yet |
| The YAML limits on `glassops.yaml` | Not yet |
| The license scan, `THIRD_PARTY_NOTICES` and the release tag rules below | Not yet |
| [Sigstore signing](sigstore-signing.md) of PERMITs, with `policy.permit` | Not yet (v1.1.0) |
| [Opt-in Rego rules](rego-rules.md), with `policy.rego` | Not yet (v1.1.0) |
| [Engine hardis](hardis-engine.md), experimental, with `salesforce.engine: hardis` | Not yet (v1.1.0) |
| [The pull request comment](pr-comment-and-telemetry.md#pull-request-comment), with `report --pr-comment` and `report.pr_comment` | Not yet (v1.1.0) |
| [OpenTelemetry export](pr-comment-and-telemetry.md#opentelemetry-export), with `telemetry.otlp` | Not yet (v1.1.0) |
| `cosign` and `opa` in both images | Not yet (v1.1.0) |

## Release tags

A release starts when a tag such as `v1.2.3` is pushed. The release workflow only starts for tags that begin like `v1.2.3`. `scripts/release/tags.sh` accepts `vMAJOR.MINOR.PATCH` with an optional suffix such as `-rc1`, and exits 2 for anything else, such as `v1.2.3.4`, so the plan job fails and nothing is published. It decides what a tag publishes, and `scripts/release/test-tags.sh` tests those rules.

| Tag | Kind | Image tags | GitHub release | Git tag moved |
|---|---|---|---|---|
| `v1.2.3-rc1` | prerelease | `1.2.3-rc1`, `v1.2.3-rc1` | marked as a prerelease | none |
| `v0.4.0` | prerelease, because the major version is 0 | `0.4.0`, `v0.4.0` | marked as a prerelease | none |
| `v1.2.3` | final | `1.2.3`, `v1.2.3`, `v1`, `latest` | a draft you publish yourself | `v1` moves to the release commit |

Details:

- The binaries go to a draft GitHub release. Publishing it stays a deliberate step.
- The images `ghcr.io/glassops-platform/glassops` and `ghcr.io/glassops-platform/glassops-analyzer` are built for `linux/amd64` and `linux/arm64`.
- The moving tags `v1` and `latest` change only on a final release, so a prerelease never becomes what `@v1` runs.
- The `v1` git tag moves last, after both images and the binaries exist. A caller of `@v1` never gets a workflow whose image is missing.
- The images carry the OCI labels `source`, `version`, `revision` and `licenses`.

### Which image a workflow runs

The reusable workflows and the `glassops-platform/glassops` action choose their default image from the tag they were called at. A call to `@v1` runs the image of the current `v1.x.y` release, and `@v1.2.3` runs the image of `v1.2.3`. When the reference is a branch or a commit, the image is `latest`. Pass the `image` input to override this. For a run you must reproduce exactly, pin a full version or an image digest.

### Tools in the images

Both images, built from this repository, hold two more programs that GlassOps runs as separate commands:

| Program | Version | Used for |
|---|---|---|
| `cosign` | v2.5.3 | [Sigstore signing](sigstore-signing.md) of PERMITs, when `policy.permit.signing` is `sigstore` |
| `opa` | 1.22.0, the static build | [Rego rules](rego-rules.md), when `policy.rego` is set |

Both are Apache-2.0 programs copied from their official images, which are pinned by digest in the Dockerfiles. They are not linked into `glassops`, and they do nothing unless you turn those features on. They are listed in `THIRD_PARTY_NOTICES` like every other package in the images. The image `ghcr.io/glassops-platform/glassops:0.4.0` does not hold them. The hardis engine's `sfdx-hardis` is in no image, see [Engine hardis](hardis-engine.md).

## License scan

Every pull request, every push to `main` and every release runs the License Scan workflow. It builds both images, writes a software bill of materials for each with syft, and fails when an image holds a package under the AGPL, GPL or LGPL that the allowlist does not list. The release workflow publishes no image unless the scan passes.

How a package is judged:

- The scan reads the license expression each package declares.
- In `A OR B`, one free side is enough, because the user may choose it. In `A AND B`, both sides must be free.
- An allowlist line is `type:name-glob = license text`. It matches only a package whose declared expression equals the text. A package whose license changes comes back into the report.
- AGPL can never be allowlisted.

The allowlist is [`scripts/notices/allowlist.txt`](https://github.com/glassops-platform/glassops/blob/main/scripts/notices/allowlist.txt). It lists the Java runtime of the analyzer image (GPLv2 with the Classpath Exception), and the Alpine base layer packages that the images contain and that GlassOps runs as separate programs, such as `git`, `busybox` and `coreutils`. They are not linked into `glassops`. The list is reviewed whenever the base image changes.

The scan is an ordinary check and not a required one. Making it required is a branch protection setting.

Run it yourself against an image you built:

```bash
scripts/notices/sbom.sh glassops:local glassops.sbom.json
python3 scripts/notices/licenses.py check glassops.sbom.json --allow scripts/notices/allowlist.txt
```

The second command exits 1 when it finds a package the allowlist does not cover.

## Third party notices

[`THIRD_PARTY_NOTICES`](https://github.com/glassops-platform/glassops/blob/main/THIRD_PARTY_NOTICES) lists the packages in both images with their licenses. The same scan workflow checks that the file still matches what the images contain, and fails when a package was added or removed and the file was not regenerated.

The notices and the project license are inside the images at `/usr/share/doc/glassops/`, and the release archives include `THIRD_PARTY_NOTICES` next to `LICENSE` and `README.md`:

```bash
docker run --rm --entrypoint cat ghcr.io/glassops-platform/glassops:v1 /usr/share/doc/glassops/THIRD_PARTY_NOTICES
```

## Tested here, and not

The tag rules, the license check and the planted-package test run in the workflow `license-scan.yml`. Pushing a real release tag has not happened with these rules yet, so the sequence of the release workflow on GitHub is untested on a real tag.
