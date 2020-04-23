---
name: Bug report
about: Tell us about a problem you are experiencing

---

**What steps did you take and what happened:**
[A clear and concise description of what the bug is, and what commands you ran.)


**What did you expect to happen:**


**The output of the following commands will help us better understand what's going on**:
(Pasting long output into a [GitHub gist](https://gist.github.com) or other pastebin is fine.)

* `kubectl logs <jiva-operator pod name> -n openebs` (optional)
* `kubectl get jv <jiva volume cr name> -n openebs -o yaml`
* `kubectl get jvp <jiva volume policy> -n openebs -o yaml`

**Anything else you would like to add:**
[Miscellaneous information that will assist in solving the issue.]


**Environment:**
- Jiva version
- OpenEBS version
- Kubernetes version (use `kubectl version`):
- Kubernetes installer & version:
- Cloud provider or hardware configuration:
- OS (e.g. from `/etc/os-release`):
