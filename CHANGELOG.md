v3.2.0 / 2022-04-19
===================
* rename 'defaultClass' helm chart template values object to 'storageClass' ([#184](https://github.com/openebs/jiva-operator/pull/184)), [@niladrih](https://github.com/niladrih))

v3.1.0 / 2022-01-03
========================
* feat(api): adding v1 CRDs ([#167](https://github.com/openebs/jiva-operator/pull/167),[@shubham14bajpai](https://github.com/shubham14bajpai))
* refactor(operator): use a common label (jivaapi) to reference jiva api ([#171](https://github.com/openebs/jiva-operator/pull/171),[@abhisheksinghbaghel](https://github.com/abhisheksinghbaghel))
* fix(provisioning): corrected the prometheus scrape annotation syntax ([#174](https://github.com/openebs/jiva-operator/pull/174),[@shazadbrohi](https://github.com/shazadbrohi))
* fix(helm): Make log verbosity configurable ([#161](https://github.com/openebs/jiva-operator/pull/161),[@ianroberts](https://github.com/ianroberts))


v3.1.0-RC2 / 2021-12-29
========================


v3.1.0-RC1 / 2021-12-20
========================
* feat(api): adding v1 CRDs ([#167](https://github.com/openebs/jiva-operator/pull/167),[@shubham14bajpai](https://github.com/shubham14bajpai))
* refactor(operator): use a common label (jivaapi) to reference jiva api ([#171](https://github.com/openebs/jiva-operator/pull/171),[@abhisheksinghbaghel](https://github.com/abhisheksinghbaghel))
* fix(provisioning): corrected the prometheus scrape annotation syntax ([#174](https://github.com/openebs/jiva-operator/pull/174),[@shazadbrohi](https://github.com/shazadbrohi))
* fix(helm): Make log verbosity configurable ([#161](https://github.com/openebs/jiva-operator/pull/161),[@ianroberts](https://github.com/ianroberts))


v3.0.0 / 2021-09-17
========================
* chore(analytic): send install event on jiva-csi controller start ([#153](https://github.com/openebs/jiva-operator/pull/153),[@mittachaitu](https://github.com/mittachaitu))

v2.12.2 / 2021-08-31
========================
* chore(build): bump csi sidecars and jiva version to 1.12.2 (293bc51, @prateekpandey14)
* refactor(e2e): migrate jiva e2e tests from openebs/e2e-test to this repo. (@nsathyaseelan)
* fix(crd): update JivaVolumePolicy parameter(monitor, enableBufio and autoScaling)  (#130, @rajaSahil)
* feat(policy): add pod AntiAffinity in jiva replica sts using policy (#132, @prateekpandey14)

v2.11.0 / 2021-07-15
========================
* fix(status): fetch volume status using controller podIP (#112, @shubham14bajpai)
* fix(csi): prevent volume mount on multiple nodes simultaneously (#107, @shubham14bajpai)

v2.11.0-RC2 / 2021-07-13
========================
* fix(status): fetch volume status using controller podIP ([#112](https://github.com/openebs/jiva-operator/pull/112),[@shubham14bajpai](https://github.com/shubham14bajpai))


v2.11.0-RC1 / 2021-07-07
========================
* fix(csi): prevent volume mount on multiple nodes simultaneously ([#107](https://github.com/openebs/jiva-operator/pull/107),[@shubham14bajpai](https://github.com/shubham14bajpai))


v2.10.0 / 2021-06-14
========================
* feat(charts): add default policy and storageclass ([#95](https://github.com/openebs/jiva-operator/pull/95),[@shubham14bajpai](https://github.com/shubham14bajpai))
* feat(operator): automate movement replicas when node is removed from cluster ([#97](https://github.com/openebs/jiva-operator/pull/97),[@shubham14bajpai](https://github.com/shubham14bajpai))


v2.10.0-RC2 / 2021-06-11
========================


v2.10.0-RC1 / 2021-06-08
========================
* feat(charts): add default policy and storageclass ([#95](https://github.com/openebs/jiva-operator/pull/95),[@shubham14bajpai](https://github.com/shubham14bajpai))
* feat(operator): automate movement replicas when node is removed from cluster ([#97](https://github.com/openebs/jiva-operator/pull/97),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.9.0 / 2021-05-15
========================
* fix(operator): use configurable serviceAccount ([#88](https://github.com/openebs/jiva-operator/pull/88),[@shubham14bajpai](https://github.com/shubham14bajpai))


v2.9.0-RC2 / 2021-05-11
========================
* fix(operator): use configurable serviceAccount ([#88](https://github.com/openebs/jiva-operator/pull/88),[@shubham14bajpai](https://github.com/shubham14bajpai))


v2.9.0-RC1 / 2021-05-06
========================


v2.8.0 / 2021-04-14
========================

* fix(controller): fix node selector for replica sts ([#65](https://github.com//pull/65),[@shubham14bajpai](https://github.com/shubham14bajpai))
* chore(k8s): updated csi driver version to v1 ([#64](https://github.com//pull/64),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.8.0-RC2 / 2021-04-12
========================

v2.8.0-RC1 / 2021-04-07
========================

* fix(controller): fix node selector for replica sts ([#65](https://github.com//pull/65),[@shubham14bajpai](https://github.com/shubham14bajpai))
* chore(k8s): updated csi driver version to v1 ([#64](https://github.com//pull/64),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.7.0 / 2021-03-16
========================

 * feat(analytics): add google analytics for jiva csi volumes ([#49](https://github.com//pull/49),[@prateekpandey14](https://github.com/prateekpandey14))
 * feat(operator): add ability to scale up replicas via JivaVolume resource ([#54](https://github.com//pull/54),[@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(build): add missing docker login and set tag stages ([#60](https://github.com//pull/60),[@prateekpandey14](https://github.com/prateekpandey14))
 * fix(resize): respond success if volume is already of same size ([#57](https://github.com//pull/57),[@payes](https://github.com/payes))
 * fix(operator): add serviceAccountName to replica sts to set the ownerreference permission ([#51](https://github.com//pull/51),[@payes](https://github.com/payes))
 * refact(operator): move operator to latest operator-sdk version ([#48](https://github.com//pull/48),[@shubham14bajpai](https://github.com/shubham14bajpai))
 * feat(operator): add events to the jivavolume controller ([#56](https://github.com//pull/56),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.7.0-RC2 / 2021-03-11
========================

 * feat(operator): add ability to scale up replicas via JivaVolume resource ([#54](https://github.com//pull/54),[@shubham14bajpai](https://github.com/shubham14bajpai))
 * fix(build): add missing docker login and set tag stages ([#60](https://github.com//pull/60),[@prateekpandey14](https://github.com/prateekpandey14))
 * fix(resize): respond success if volume is already of same size ([#57](https://github.com//pull/57),[@payes](https://github.com/payes))
 * feat(operator): add events to the jivavolume controller ([#56](https://github.com//pull/56),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.7.0-RC1 / 2021-03-09
========================

 * fix(operator): add serviceAccountName to replica sts to set the ownerreference permission ([#51](https://github.com//pull/51),[@payes](https://github.com/payes))
 * feat(analytics): add google analytics for jiva csi volumes ([#49](https://github.com//pull/49),[@prateekpandey14](https://github.com/prateekpandey14))
 * refact(operator): move operator to latest operator-sdk version ([#48](https://github.com//pull/48),[@shubham14bajpai](https://github.com/shubham14bajpai))

v2.6.0 / 2021-02-15
========================

  *  chore(build): merge jiva-csi and jiva-operator repos ([#28](https://www.github.com/openebs/jiva-operator#28), [@shubham14bajpai](https://github.com/shubham14bajpai))
  *  feat(version): add support for version details ([#30](https://www.github.com/openebs/jiva-operator#30), [@shubham14bajpai](https://github.com/shubham14bajpai))
  *  feat(build): add multi arch support for csi driver and operator ([#31](https://www.github.com/openebs/jiva-operator#28), [@prateekpandey14](https://github.com/prateekpandey14))
  *  feat(policy): add support for target affinity for controller ([#32](https://www.github.com/openebs/jiva-operator#32), [@shubham14bajpai](https://github.com/shubham14bajpai))
  *  feat(raw-block-volumes): add support for raw block volumes ([#33](https://www.github.com/openebs/jiva-operator#33), [@payes](https://github.com/payes))
  *  fix(crds): update the crds with proper json tags ([#34](https://www.github.com/openebs/jiva-operator#34), [@shubham14bajpai](https://github.com/shubham14bajpai))
  *  feat(policy): add support for target affinity for controller ([#35](https://www.github.com/openebs/jiva-operator#35), [@prateekpandey14 ](https://github.com/prateekpandey14))
  *  feat(charts): set up helm chart ci and release workflow ([#37](https://www.github.com/openebs/jiva-operator#37), [@prateekpandey14](https://github.com/prateekpandey14))
  *  feat(topology): add custom topology key support for provisioning ([#39](https://www.github.com/openebs/jiva-operator#39), [@prateekpandey14 ](https://github.com/prateekpandey14))
  *   fix(reconcile): removed defer to enable reconcile on failed volumes ([#40](https://www.github.com/openebs/jiva-operator#40), [@shubham14bajpai](https://github.com/shubham14bajpai))
  *   refact(logs): Use logrus for logging ([#41](https://www.github.com/openebs/jiva-operator#41), [@payes](https://github.com/payes))

1.7.0-RC1 / 2020-02-05
========================

  *  Add schema for JivaVolumePolicy CR ([#8](https://www.github.com/openebs/jiva-operator#8), [@shubham14bajpai](https://github.com/shubham14bajpai))
  *  Add resize related schema ([#6](https://www.github.com/openebs/jiva-operator#6), [@utkarshmani1997](https://github.com/utkarshmani1997))
  *  Add target and staging path to JivaVolumeCR ([#9](https://www.github.com/openebs/jiva-operator#9), [@utkarshmani1997](https://github.com/utkarshmani1997))
