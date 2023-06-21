/*
Copyright 2018, 2019 the Velero contributors.

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
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	//"k8s.io/apimachinery/pkg/api/meta"
	//"k8s.io/apimachinery/pkg/util/yaml"
	"gopkg.in/yaml.v2"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	//"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RestorePlugin is a restore item action plugin for Velero
type RestorePlugin struct {
	log logrus.FieldLogger
}

// NewRestorePlugin instantiates a RestorePlugin.
func NewRestorePlugin(log logrus.FieldLogger) *RestorePlugin {
	return &RestorePlugin{log: log}
}

// AppliesTo returns information about which resources this action should be invoked for.
// The IncludedResources and ExcludedResources slices can include both resources
// and resources with group names. These work: "ingresses", "ingresses.extensions".
// A RestoreItemAction's Execute function will only be invoked on items that match the returned
// selector. A zero-valued ResourceSelector matches all resources.
func (p *RestorePlugin) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{
		IncludedResources: []string{"ingresses"},
	}, nil
}

// Execute allows the RestorePlugin to perform arbitrary logic with the item being restored,
// in this case, setting a custom annotation on the item being restored.
func (p *RestorePlugin) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	ingress := networkingv1beta1.Ingress{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(input.Item.UnstructuredContent(), &ingress); err != nil {
				return nil, errors.WithStack(err)
	}
	p.log.Infof("Peter: ingress APIVersion=%s",ingress.APIVersion)
	p.log.Infof("Peter: ingress all content=%s",ingress)
	if ingress.APIVersion == "networking.k8s.io/v1beta1" {
		ingressV1 := networkingv1.Ingress{}
		//p.log.Infof("Peter: EMPTY ingress v1 -- =%s",ingressV1)
		ingressV1.Kind = "Ingress"
		ingressV1.APIVersion = "networking.k8s.io/v1"
		//p.log.Infof("Peter:  After with KIND: ingress v1 -- =%s",ingressV1)
		// 复制元数据信息
		ingressV1.ObjectMeta.Name = ingress.ObjectMeta.Name
		ingressV1.ObjectMeta.Namespace = ingress.ObjectMeta.Namespace

		// 复制spec
		ingressV1.Spec.Rules = make([]networkingv1.IngressRule, len(ingress.Spec.Rules))
		for i, rule := range ingress.Spec.Rules {
			ingressV1.Spec.Rules[i].Host = rule.Host
			ingressV1.Spec.Rules[i].IngressRuleValue = networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: make([]networkingv1.HTTPIngressPath, len(rule.HTTP.Paths)),
				},
			}
			for j, path := range rule.HTTP.Paths {
				ingressV1.Spec.Rules[i].IngressRuleValue.HTTP.Paths[j] = networkingv1.HTTPIngressPath{
					Path: path.Path,
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: path.Backend.ServiceName,
							Port: networkingv1.ServiceBackendPort{
								Number: path.Backend.ServicePort.IntVal,
							},
						},
					},
				}
			}
		}
		p.log.Infof("Peter: ingressV1 Kind=%s, APIVersion=%s",ingressV1.Kind, ingressV1.APIVersion)
		/*if true{
			outBytes, err := yaml.Marshal(ingressV1)
			if err != nil {
				panic(err.Error())
			}
			// 输出 YAML 格式的 Ingress v1数据
			p.log.Infof("Peter The YAML is =%s",string(outBytes))
		}*/

		inputMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ingressV1)
		if err != nil {
			return nil, errors.WithStack(err)

		}
		return velero.NewRestoreItemActionExecuteOutput(&unstructured.Unstructured{Object: inputMap}), nil


	} else {
		return velero.NewRestoreItemActionExecuteOutput(input.Item), nil 
	}

	/*
	ObjectMeta, err := meta.Accessor(input.Item)
	if err != nil {
		return &velero.RestoreItemActionExecuteOutput{}, err
	}

	annotations := ObjectMeta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations["velero.io/my-restore-plugin"] = "1"

	ObjectMeta.SetAnnotations(annotations)

	return velero.NewRestoreItemActionExecuteOutput(input.Item), nil
	*/
}
