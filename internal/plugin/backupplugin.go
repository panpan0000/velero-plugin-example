/*
Copyright 2017, 2019 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"strings"
    //#"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	appsv1API "k8s.io/api/apps/v1"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
//		"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

)

// BackupPlugin is a backup item action plugin for Velero.
type BackupPlugin struct {
	log logrus.FieldLogger
}

// NewBackupPlugin instantiates a BackupPlugin.
func NewBackupPlugin(log logrus.FieldLogger) *BackupPlugin {
	return &BackupPlugin{log: log}
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A BackupPlugin's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *BackupPlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
				IncludedResources: []string{"deployments"},
					
	}, nil
	
}

// Execute allows the ItemAction to perform arbitrary logic with the item being backed up,
// in this case, setting a custom annotation on the item being backed up.
func (p *BackupPlugin) Execute(item runtime.Unstructured, backup *v1.Backup) (runtime.Unstructured, []velero.ResourceIdentifier, error) {
	p.log.Info("Hello from my BackupPlugin(v1)!")


	deployment := appsv1API.Deployment{}
	itemMarshal, _ := json.Marshal(item)
	json.Unmarshal(itemMarshal, &deployment)
	p.log.Infof("[peter] deployment: %s", deployment.Name)


	annotations := deployment.GetAnnotations()
	p.log.Infof("[peter] deployment annotations: %s", annotations)

	if annotations == nil {
		annotations = make(map[string]string)
	}

	/// find parcel in annotations
	if t, found := annotations["dce.daocloud.io/parcel.net.type"]; found {
		if t == "ovs" {
			if parcelValue, found := annotations["dce.daocloud.io/parcel.net.value"]; found {
				p.log.Infof("[peter] parcel: %s", parcelValue)
				kv := strings.Split(parcelValue, ":")
				if len(kv) == 2{
					annotations["v1.multus-cni.io/default-network"] = "kube-system/macvlan-standalone"
					annotations["ipam.spidernet.io/ippools"] = "[{\"interface\":\"eth0\", \"ipv4\":[\"" + kv[1]  + "\"]}]"
				}
			}
		}
	}


	//metadata.SetAnnotations(annotations)
	deployment.SetAnnotations(annotations)
	p.log.Infof("[peter] after SetAnnotations = : %s", deployment.GetAnnotations())


	var out map[string]interface{}
	objrec, _ := json.Marshal(deployment)
	json.Unmarshal(objrec, &out)
	item.SetUnstructuredContent(out)
	return item, nil, nil 

}
