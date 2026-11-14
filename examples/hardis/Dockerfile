# An image you build yourself to run GlassOps with salesforce.engine: hardis.
#
# GlassOps does not ship sfdx-hardis and this repository never builds or publishes this image.
# sfdx-hardis is a separate program under its own license (AGPL-3.0). Read that license before
# you build, run or share an image that contains it.
#
# Build it from the GlassOps image you already use:
#
#   docker build -f examples/hardis/Dockerfile \
#     --build-arg GLASSOPS_IMAGE=ghcr.io/glassops-platform/glassops:<version> \
#     -t my-glassops-hardis .
#
# Then run `glassops doctor` in it. Doctor checks that sfdx-hardis is installed at
# salesforce.hardis.min_version or later.

ARG GLASSOPS_IMAGE=ghcr.io/glassops-platform/glassops:latest
FROM ${GLASSOPS_IMAGE}

# The sfdx-hardis version GlassOps was tested with. Doctor warns on a newer one.
ARG HARDIS_VERSION=8.20.1
RUN echo y | sf plugins install sfdx-hardis@${HARDIS_VERSION} \
    && sf plugins --json > /dev/null \
    && npm cache clean --force
