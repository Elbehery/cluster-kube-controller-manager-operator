package clustername

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/clock"

	configv1 "github.com/openshift/api/config/v1"
	configlistersv1 "github.com/openshift/client-go/config/listers/config/v1"
	"github.com/openshift/library-go/pkg/operator/events"

	"github.com/ghodss/yaml"

	"github.com/openshift/cluster-kube-controller-manager-operator/pkg/operator/configobservation"
)

func TestObserveInfraID(t *testing.T) {
	type Test struct {
		name            string
		config          *configv1.Infrastructure
		input, expected map[string]interface{}
	}
	tests := []Test{
		{
			name: "new name, no old config",
			config: &configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
				Status: configv1.InfrastructureStatus{
					InfrastructureName: "newClusterName",
				},
			},
			input: map[string]interface{}{},
			expected: map[string]interface{}{
				"extendedArguments": map[string]interface{}{
					"cluster-name": []interface{}{
						"newClusterName",
					},
				},
			},
		},
		{
			name: "new name, old config",
			config: &configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
				Status: configv1.InfrastructureStatus{
					InfrastructureName: "newClusterName",
				},
			},
			input: map[string]interface{}{
				"extendedArguments": map[string]interface{}{
					"cluster-name": []interface{}{
						"oldClusterName",
					},
				},
			},
			expected: map[string]interface{}{
				"extendedArguments": map[string]interface{}{
					"cluster-name": []interface{}{
						"newClusterName",
					},
				},
			},
		},
		{
			name:     "none, no old config",
			config:   &configv1.Infrastructure{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}},
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:   "none, existing config",
			config: &configv1.Infrastructure{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}},
			input: map[string]interface{}{
				"extendedArguments": map[string]interface{}{
					"cluster-name": []interface{}{
						"oldClusterName",
					},
				},
			},
			expected: map[string]interface{}{
				"extendedArguments": map[string]interface{}{
					"cluster-name": []interface{}{
						"oldClusterName",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			if err := indexer.Add(test.config); err != nil {
				t.Fatal(err.Error())
			}
			listers := configobservation.Listers{
				InfrastructureLister_: configlistersv1.NewInfrastructureLister(indexer),
			}
			result, errs := ObserveInfraID(listers, events.NewInMemoryRecorder("infraid", clock.RealClock{}), test.input)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else {
				if !reflect.DeepEqual(test.expected, result) {
					t.Errorf("\n===== observed config expected:\n%v\n===== observed config actual:\n%v", toYAML(test.expected), toYAML(result))
				}
			}
		})
	}
}

func toYAML(o interface{}) string {
	b, e := yaml.Marshal(o)
	if e != nil {
		return e.Error()
	}
	return string(b)
}
