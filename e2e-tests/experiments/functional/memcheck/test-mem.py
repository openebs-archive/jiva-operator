# -*- coding: utf-8 -*-
#!/usr/bin/env python
#description     :Test to verify the memory consumed with sample workloads. 
#==============================================================================

# Copyright © 2019-2020 The OpenEBS Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# You may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import division
import subprocess
import time, os, sys
list = []
namespace = sys.argv[1]
benchmark = sys.argv[2]
pvname = sys.argv[3]
cmd_cntrl_name = "kubectl get pod -n %s -l openebs.io/component=jiva-controller,openebs.io/persistent-volume=%s --no-headers | awk '{print $1}'" %(namespace,pvname)
print cmd_cntrl_name
out = subprocess.Popen(cmd_cntrl_name,stdout=subprocess.PIPE,shell=True)
cntrl_name = out.communicate()
cntrl_pod_name = cntrl_name[0].strip('\n')
n = cntrl_pod_name.split('-')
lst = n[:len(n)-2]
lst.append("con")
container_name = "-".join(lst)
print container_name
used_mem_process = "kubectl exec %s -n %s -- pmap -x 1 | awk ''/total'/ {print $3}'" %(cntrl_pod_name,namespace)
print used_mem_process
n = 10
count = 0
#Obtaining memory consumed by longhorn process from the cntroller pod.
while count < n:
    count = count + 1
    out = subprocess.Popen(used_mem_process,stdout=subprocess.PIPE,shell=True)
    used_mem = out.communicate()
    mem_in_mb = int(used_mem[0])/1024
    print mem_in_mb, "MB"
    if mem_in_mb < benchmark:
        time.sleep(20)
    else:
        print "Fail"
        #break
        quit()
    list.append(mem_in_mb)
print list
# A watermark of 800MB(re-calibrated based on results oberved from latest sanity run) 
# profile chosen in this test
# TODO: Identify better mem consumption strategies
if all(i <= benchmark for i in list):
        print "Pass"
else:
        print "Fail"

