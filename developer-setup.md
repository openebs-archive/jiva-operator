# Development Workflow

## Prerequisites

* You have Go 1.14 installed on your local host/development machine.
* You have Docker installed on your local host/development machine. Docker is required for building jiva-operator container images and to push them into a Kubernetes cluster for testing.
* You have `kubectl` installed. For running integration tests, you will require atleast an single node Kubernetes cluster. If you are using `minikube`, use with `vm-driver=none`. Don't worry if you don't have access to the Kubernetes cluster, raising a PR with the jiva-operator repository will run integration tests for your changes against a Minikube cluster.

## Initial Setup

### Fork in the cloud

1. Visit https://github.com/openebs/jiva-operator
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone fork to local host

Place openebs/jiva-operator's code on your `GOPATH` using the following cloning procedure.
Create your clone:

```
mkdir -p $GOPATH/src/github.com/openebs
cd $GOPATH/src/github.com/openebs

# Note: Here user= your github profile name
git clone https://github.com/$user/jiva-operator.git

# Configure remote upstream
cd $GOPATH/src/github.com/openebs/jiva-operator
git remote add upstream https://github.com/openebs/jiva-operator.git

# Never push to upstream develop
git remote set-url --push upstream no_push

# Confirm that your remotes make sense:
git remote -v
```
> **Note:** If your `GOPATH` has more than one (`:` separated) paths in it, then you should use *one of your go path* instead of `$GOPATH` in the commands mentioned here. This statement holds throughout this document.

## Building and Testing your changes

* To build the jiva-operator binary
  ```
  go build
  ```

* To build the docker image
  ```
  make build
  ```

* To build operator specific changes
 ```
 cd $GOPATH/src/github.com/openebs/jiva-operator
 make build.operator
 ```

* To build CSI driver plugin specific changes
 ```
 cd $GOPATH/src/github.com/openebs/jiva-operator
 make build.plugin
 ```

* Test your changes
  Integration tests are written using Ginkgo under [tests](./tests/) and jiva-operator controller and replicas are run as docker containers.
  To run the run the integration tests locally, run
  ```
  cd $GOPATH/src/github.com/openebs/jiva-operator
  kubectl apply -f https://openebs.github.io/charts/hostpath-operator.yaml
  kubectl apply -f deploy/hostpath-sc.yaml
  kubectl apply -f deploy/operator.yaml
  kubectl apply -f deploy/jiva-csi.yaml
  ./ci/ci.sh
  cd ./tests
  make tests
  ```

## Commit your changes

The commits should follow the [code-standards](code-standard.md).

## Git Development Workflow

### Always sync your local repository:
Open a terminal on your local host. Change directory to the jiva-operator fork root.

```
$ cd $GOPATH/src/github.com/openebs/jiva-operator
```

Checkout the develop branch.

```
$ git checkout develop
Switched to branch 'develop'
Your branch is up-to-date with 'origin/develop'.
```

Recall that origin/develop is a branch on your remote GitHub repository.
Make sure you have the upstream remote openebs/jiva-operator by listing them.

 ```
 $ git remote -v
 origin	https://github.com/$user/jiva-operator.git (fetch)
 origin	https://github.com/$user/jiva-operator.git (push)
 upstream	https://github.com/openebs/jiva-operator.git (fetch)
 upstream	https://github.com/openebs/jiva-operator.git (no_push)
 ```

 If the upstream is missing, add it by using below command.

 ```
 $ git remote add upstream https://github.com/openebs/jiva-operator.git
 ```
 Fetch all the changes from the upstream develop branch.

 ```
 $ git fetch upstream develop
 remote: Counting objects: 141, done.
 remote: Compressing objects: 100% (29/29), done.
 remote: Total 141 (delta 52), reused 46 (delta 46), pack-reused 66
 Receiving objects: 100% (141/141), 112.43 KiB | 0 bytes/s, done.
 Resolving deltas: 100% (79/79), done.
 From github.com:openebs/jiva-operator
   * branch            develop     -> FETCH_HEAD
 ```

 Rebase your local develop with the upstream/develop.

 ```
 $ git rebase upstream/develop
 First, rewinding head to replay your work on top of it...
 Fast-forwarded develop to upstream/develop.
 ```
 This command applies all the commits from the upstream develop to your local develop.

 Check the status of your local branch.

 ```
 $ git status
 On branch develop
 Your branch is ahead of 'origin/develop' by 12 commits.
 (use "git push" to publish your local commits)
 nothing to commit, working directory clean
 ```
 Your local repository now has all the changes from the upstream remote. You need to push the changes to your own remote fork which is origin develop.

 Push the rebased develop to origin develop.

 ```
 $ git push origin develop
 Username for 'https://github.com': $user
 Password for 'https://$user@github.com':
 Counting objects: 223, done.
 Compressing objects: 100% (38/38), done.
 Writing objects: 100% (69/69), 8.76 KiB | 0 bytes/s, done.
 Total 69 (delta 53), reused 47 (delta 31)
 To https://github.com/$user/jiva-operator.git
 8e107a9..5035fa1  develop -> develop
 ```

### Contributing to a feature or bugfix.

Always start with creating a new branch from develop to work on a new feature or bugfix. Your branch name should have the format XX-descriptive where XX is the issue number you are working on followed by some descriptive text. For example:

 ```
 $ git checkout develop
 # Make sure the develop is rebased with the latest changes as described in previous step.
 $ git checkout -b 1234-fix-developer-docs
 Switched to a new branch '1234-fix-developer-docs'
 ```
Happy Hacking!

### Keep your branch in sync

[Rebasing](https://git-scm.com/docs/git-rebase) is very import to keep your branch in sync with the changes being made by others and to avoid huge merge conflicts while raising your Pull Requests. You will always have to rebase before raising the PR.

```
# While on your myfeature branch (see above)
git fetch upstream
git rebase upstream/develop
```

While you rebase your changes, you must resolve any conflicts that might arise and build and test your changes using the above steps.

## Submission

### Create a pull request

Before you raise the Pull Requests, ensure you have reviewed the checklist in the [CONTRIBUTING GUIDE](../CONTRIBUTING.md):
- Ensure that you have re-based your changes with the upstream using the steps above.
- Ensure that you have added the required unit tests for the bug fixes or new feature that you have introduced.
- Ensure your commits history is clean with proper header and descriptions.

Go to the [openebs/jiva-operator github](https://github.com/openebs/jiva-operator) and follow the Open Pull Request link to raise your PR from your development branch.

