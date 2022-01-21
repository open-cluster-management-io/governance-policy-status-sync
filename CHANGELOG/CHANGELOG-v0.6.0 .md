# v0.6.0

* Updated the operator-sdk from v0.x to v1.x. This is mostly an internal change,
  however, the controller now uses both `ConfigMap`s and the `Lease` API for
  leader election by default. If this controller is running on a Kubernetes
  version without the `Lease` API, use the `--legacy-leader-elect` flag.
* Automatically restart the pod when the Hub kubeconfig changes.
* Handle Policy status updates when a managed cluster resumes from hibernation.
