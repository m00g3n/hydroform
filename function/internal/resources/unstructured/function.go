package unstructured

import (
	"io/ioutil"
	"path"

	"github.com/kyma-incubator/hydroform/function/internal/resources/types"
	"github.com/kyma-incubator/hydroform/function/internal/workspace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	functionApiVersion = "serverless.kyma-project.io/v1alpha1"
)

func NewFunction(cfg workspace.Cfg, refs ...map[string]interface{}) (unstructured.Unstructured, error) {
	return newFunction(cfg, ioutil.ReadFile, refs...)
}

func newFunction(cfg workspace.Cfg, readFile ReadFile, refs ...map[string]interface{}) (unstructured.Unstructured, error) {
	out := unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": functionApiVersion,
		"kind":       "Function",
		"metadata": map[string]interface{}{
			"name":            cfg.Name,
			"labels":          cfg.Labels,
			"ownerReferences": refs,
		},
		"spec": map[string]interface{}{
			"runtime": cfg.Runtime,
		},
	}}

	spec := out.Object["spec"].(map[string]interface{})
	for key, value := range runtimeMappings[cfg.Runtime] {
		filePath := path.Join(cfg.SourcePath, string(value))
		data, err := readFile(filePath)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
		if len(data) == 0 {
			continue
		}
		spec[string(key)] = string(data)
	}

	var resources map[string]interface{}
	if cfg.Resources.Requests != nil {
		resources = map[string]interface{}{}
		resources["requests"] = cfg.Resources.Requests
	}
	if cfg.Resources.Limits != nil {
		if resources == nil {
			resources = map[string]interface{}{}
		}
		resources["limits"] = cfg.Resources.Limits
	}
	if resources != nil {
		spec["resources"] = resources
	}
	if len(refs) == 0 {
		return out, nil
	}

	return out, nil
}

type property string

const (
	propertySource property = "source"
	propertyDeps   property = "deps"
)

var (
	runtimeMappings = map[types.Runtime]map[property]workspace.FileName{
		types.Nodejs12: {
			propertySource: workspace.FileNameHandlerJs,
			propertyDeps:   workspace.FileNamePackageJSON,
		},
		types.Nodejs10: {
			propertySource: workspace.FileNameHandlerJs,
			propertyDeps:   workspace.FileNamePackageJSON,
		},
		types.Python38: {
			propertySource: workspace.FileNameHandlerPy,
			propertyDeps:   workspace.FileNameRequirementsTxt,
		},
	}
)
