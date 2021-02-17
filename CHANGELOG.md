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
