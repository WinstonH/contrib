# Packer Builder

 Use Packer from https://www.packer.io/ to build bootable Kubernetes images for
 multiple cloud providers

```
 Packer is a tool for creating machine and container images for multiple
 platforms from a single source configuration.
```

## Goal

The goal of this builder is to build a bootable image for various machine image
types (GCE/VirtualBox/VMWare/AWS et al) given the Git refpoint.

## Dockerized Build environment

To eliminate dependencies on the build environment, this build will be run in a
pre-built Docker container with all required dependencies installed.
