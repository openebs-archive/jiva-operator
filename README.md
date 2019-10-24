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

Jiva operator usage Operator SDK framework for the development and does following
- Introduce a new Custom Resource for Jiva Volume for managing the lifecycle of
  jiva volumes
- Jiva Operator that watches and operates on the Jiva Volume CR to handle
  degraded cases where the node are permanently gone or to handle scale up/down
  of replicas
- Support for Resize of Jiva Volumes

## Quick Start




