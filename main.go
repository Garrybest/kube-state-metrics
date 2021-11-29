/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	samplev1alpha1 "k8s.io/sample-controller/pkg/apis/samplecontroller/v1alpha1"
	clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"

	"k8s.io/kube-state-metrics/v2/pkg/app"
	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

func main() {
	app.RunKubeStateMetrics(new(FooFactory))
}

var (
	descFooLabelsDefaultLabels = []string{"namespace", "foo"}
)

type FooFactory struct{}

func (f *FooFactory) Name() string {
	return "foos"
}

func (f *FooFactory) CreateClient(cfg *rest.Config) (interface{}, error) {
	return clientset.NewForConfig(cfg)
}

func (f *FooFactory) MetricFamilyGenerators(allowAnnotationsList, allowLabelsList []string) []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kube_foo_spec_replicas",
			"Number of desired replicas for a foo.",
			metric.Gauge,
			"",
			wrapFooFunc(func(f *samplev1alpha1.Foo) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(*f.Spec.Replicas),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kube_foo_status_replicas_available",
			"The number of available replicas per foo.",
			metric.Gauge,
			"",
			wrapFooFunc(func(f *samplev1alpha1.Foo) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							Value: float64(f.Status.AvailableReplicas),
						},
					},
				}
			}),
		),
	}
}

func wrapFooFunc(f func(*samplev1alpha1.Foo) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		foo := obj.(*samplev1alpha1.Foo)

		metricFamily := f(foo)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descFooLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{foo.Namespace, foo.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func (f *FooFactory) ExpectedType() interface{} {
	return &samplev1alpha1.Foo{}
}

func (f *FooFactory) ListWatch(customResourceClient interface{}, ns string, fieldSelector string) cache.ListerWatcher {
	client := customResourceClient.(*clientset.Clientset)
	return &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			opts.FieldSelector = fieldSelector
			return client.SamplecontrollerV1alpha1().Foos(ns).List(context.Background(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			opts.FieldSelector = fieldSelector
			return client.SamplecontrollerV1alpha1().Foos(ns).Watch(context.Background(), opts)
		},
	}
}
