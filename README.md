# jiva-operator

[![Releases](https://img.shields.io/github/release/openebs/openebs/all.svg?style=flat-square)](https://github.com/openebs/openebs/releases)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)]( https://openebs-community.slack.com)
[![Twitter](https://img.shields.io/twitter/follow/openebs.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=openebs)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/openebs/openebs/blob/master/CONTRIBUTING.md)


https://openebs.org/

## Overview

Jiva operator is a custom kubernetes controller, which will continuously watch
for the JivaVolume CR and will do the bootstrapping of jiva components such as
creating service and deployment of jiva controller, deploy jiva replicas as
statefulsets using localpv for persisting the data.

Jiva Operator helps with managing the lifecycle and operations on Jiva Volumes.
Jiva Operator use the JivaVolume CR to perform the operations on Jiva Volume and it is implemented using Operator SDK.

Jiva Operator does the following:

- Launches Jiva Volume CRD into the cluster if not already present.
- Launches Jiva Volume components, when a new Jiva Volume CR is created and updates the Jiva Volume CR status attributes.
- Clears the Jiva Volume components when a Jiva Volume CR is deleted
- Performs update operations like Volume Expansion
- Handles scenarios like node failure and creation of new replicas.

A Jiva Volume comprises of the following components:

- Jiva Target Deployment and an associated Service
- Jiva Replica StatefulSet with Hostpath Local PVs for saving the data.

## Quick Start




