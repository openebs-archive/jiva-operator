# Jiva Operator

[![Build Status](https://github.com/openebs/jiva-operator/actions/workflows/build.yaml/badge.svg)](https://github.com/openebs/jiva-operator/actions/workflows/build.yml)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/hiRcXyDTRVO2_Zs9fp0CAg)
[![Twitter](https://img.shields.io/twitter/follow/openebs.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=openebs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/openebs/openebs/blob/HEAD/CONTRIBUTING.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/openebs/jiva-csi)](https://goreportcard.com/report/github.com/openebs/jiva-operator)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fjiva-operator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fjiva-operator?ref=badge_shield)
[![Releases](https://img.shields.io/github/release/openebs/jiva-operator/all.svg?style=flat-square)](https://github.com/openebs/jiva-operator/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/openebs/jiva)](https://hub.docker.com/repository/docker/openebs/jiva)

<img width="200" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/HEAD/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

<p align="justify">
<strong>OpenEBS Jiva Operator</strong> can be used to dynamically provision highly available Kubernetes Persistent Volumes using local (ephemeral) storage available on the Kubernetes nodes. 
</p>
<br>

### Overview

Jiva Operator helps with managing the lifecycle and operations on Jiva Volumes. Jiva Volumes are highly available block volumes that save the data to local storage available on the Kubernetes nodes. Jiva volumes replicate the volume data to multiple nodes to provide resiliency against node failures.

A Jiva Volume comprises of the following Kubernetes components:

- Jiva Controller Deployment and an associated Service.
- Jiva Replica StatefulSet with Hostpath Local PVs for saving the data.
- JivaVolume CR containing the status of the volume.

We are always happy to list users who run Jiva in production, check out our existing [adopters](./ADOPTERS.md).

## Compatibility 

| K8s Version | Jiva Version     | Project Status
| ------------| ---------------- | ------------- 
| 1.18+       | 2.6.0+           | Beta  
| 1.20+       | 2.12.0+          | Beta  

## Usage

- [Quickstart guide](docs/quickstart.md)


## Supported Features

- [x] Dynamic provisioning
- [x] Enforced volume size limit
- [x] Thin provisioned
- [x] High Availability
- [x] Access Modes
    - [x] ReadWriteOnce
    - [x] ReadWriteMany (using NFS)
    - ~~ReadOnlyMany~~
- [x] Volume modes
    - [x] `Filesystem` mode
    - [x] `Block` mode
- [x] Volume metrics
- [x] Supports fsTypes: `ext4`, `btrfs`, `xfs`
- [x] Online expansion: If fs supports it (e.g. ext4, btrfs, xfs)
- [x] Backup and Restore (using Velero)
- [x] Supports OS/ARCH: linux/arm64, linux/amd64


### Contributing

OpenEBS welcomes your feedback and contributions in any form possible.

- [Join OpenEBS community on Kubernetes Slack](https://kubernetes.slack.com)
  - Already signed up? Head to our discussions at [#openebs](https://kubernetes.slack.com/messages/openebs/)
- Want to raise an issue or help with fixes and features?
  - See [open issues](https://github.com/openebs/jiva-operator/issues)
  - See [contributing guide](./CONTRIBUTING.md)
  - See [Project Roadmap](https://github.com/openebs/openebs/blob/HEAD/ROADMAP.md#jiva)
  - Want to join our contributor community meetings, [check this out](https://hackmd.io/mfG78r7MS86oMx8oyaV8Iw?view).
- Join our OpenEBS CNCF Mailing lists
  - For OpenEBS project updates, subscribe to [OpenEBS Announcements](https://lists.cncf.io/g/cncf-openebs-announcements)
  - For interacting with other OpenEBS users, subscribe to [OpenEBS Users](https://lists.cncf.io/g/cncf-openebs-users)
### Code of conduct

Please read the community code of conduct [here](./CODE_OF_CONDUCT.md).

[Docker environment]: https://docs.docker.com/engine
[Go environment]: https://golang.org/doc/install
[openebs/openebs]: https://github.com/openebs/openebs
[channel]: https://kubernetes.slack.com/messages/openebs/


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fjiva-operator.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fjiva-operator?ref=badge_large)
