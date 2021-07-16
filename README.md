# Jiva Operator

[![Releases](https://img.shields.io/github/release/openebs/openebs/all.svg?style=flat-square)](https://github.com/openebs/openebs/releases)
[![Build Status](https://github.com/openebs/jiva-operator/actions/workflows/build.yaml/badge.svg)](https://github.com/openebs/jiva-operator/actions/workflows/build.yml)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/hiRcXyDTRVO2_Zs9fp0CAg)
[![Twitter](https://img.shields.io/twitter/follow/openebs.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=openebs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/openebs/openebs/blob/master/CONTRIBUTING.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/openebs/jiva-csi)](https://goreportcard.com/report/github.com/openebs/jiva-operator)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fjiva-operator.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fjiva-operator?ref=badge_shield)

### Overview

Jiva Operator helps with managing the lifecycle and operations on Jiva Volumes.

A Jiva Volume comprises of the following components:

- Jiva Controller Deployment and an associated Service.
- Jiva Replica StatefulSet with Hostpath Local PVs for saving the data.
- JivaVolume CR containing the status of the volume.

We are always happy to list users who run Jiva in production, check out our existing [adopters](https://github.com/openebs/openebs/tree/master/adopters), and their [feedbacks](https://github.com/openebs/openebs/issues/2719).

### Compatibility and Feature Matrix

| Project Status | Operator Version | K8s Version | OpenEBS Version | Dynamic Provisioning | Resize (Expansion) | Snapshots | Raw Block | AccessModes |
| ---------------- | --------------- | ------------------- | --------------- | --------------------------- | ----------- | --------- | --------- | ---------- |
| Beta | 2.6.0+ |   1.18+   |   2.6.0+   |   yes   |    yes    |   no   |   yes   |   RWO   |



## Usage

- [Quickly deploy it on K8s and get started](docs/quickstart.md)
- [Policies Tutorial](docs/tutorials/policies.md)
- [Troubleshooting Guide](https://docs.openebs.io/docs/next/t-jiva.html)
- [Monitoring](https://github.com/openebs/monitoring/blob/develop/docs/metrics-jiva.md)

### Raising Issues And PRs

If you want to raise any issue for jiva-operator please do that at [openebs/openebs].

### Contributing

OpenEBS welcomes your feedback and contributions in any form possible.

- [Join OpenEBS community on Kubernetes Slack](https://kubernetes.slack.com)
  - Already signed up? Head to our discussions at [#openebs](https://kubernetes.slack.com/messages/openebs/)
- Want to raise an issue or help with fixes and features?
  - See [open issues](https://github.com/openebs/openebs/issues)
  - See [contributing guide](./CONTRIBUTING.md)
  - See [Project Roadmap](https://github.com/openebs/openebs/blob/master/ROADMAP.md#jiva)
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