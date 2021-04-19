# Jiva Operator

[![Releases](https://img.shields.io/github/release/openebs/openebs/all.svg?style=flat-square)](https://github.com/openebs/openebs/releases)
[![Slack](https://img.shields.io/badge/JOIN-SLACK-blue)](https://kubernetes.slack.com/messages/openebs/)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/hiRcXyDTRVO2_Zs9fp0CAg)
[![Twitter](https://img.shields.io/twitter/follow/openebs.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=openebs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/openebs/openebs/blob/master/CONTRIBUTING.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/openebs/jiva-csi)](https://goreportcard.com/report/github.com/openebs/jiva-operator)

### Overview

Jiva Operator helps with managing the lifecycle and operations on Jiva Volumes.

A Jiva Volume comprises of the following components:

- Jiva Controller Deployment and an associated Service.
- Jiva Replica StatefulSet with Hostpath Local PVs for saving the data.
- JivaVolume CR containing the status of the volume.

### Compatibility and Feature Matrix

| Project Status | Operator Version | K8s Version | OpenEBS Version | Dynamic Provisioning | Resize (Expansion) | Snapshots | Raw Block | AccessModes |
| ---------------- | --------------- | ------------------- | --------------- | --------------------------- | ----------- | --------- | --------- | ---------- |
| Beta | 2.6.0+ |   1.18+   |   2.6.0+   |   yes   |    yes    |   no   |   yes   |   RWO   |



## Usage

- [Quickly deploy it on K8s and get started](docs/quickstart.md)
- [Policies Tutorial](docs/tutorials/policies.md)

### Raising Issues And PRs

If you want to raise any issue for jiva-operator please do that at [openebs/openebs].

### Contributing

If you would like to contribute to code and are unsure about how to proceed,
please get in touch with the maintainers on Kubernetes Slack #openebs [channel].

Please read the contributing guidelines [here](./CONTRIBUTING.md).

### Code of conduct

Please read the community code of conduct [here](./CODE_OF_CONDUCT.md).

[Docker environment]: https://docs.docker.com/engine
[Go environment]: https://golang.org/doc/install
[openebs/openebs]: https://github.com/openebs/openebs
[channel]: https://kubernetes.slack.com/messages/openebs/
