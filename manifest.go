package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/geofffranks/spruce"
	"github.com/pivotal-golang/lager"
	"gopkg.in/yaml.v2"
)

func wrap(key string, data map[interface{}]interface{}) map[interface{}]interface{} {
	kk := strings.Split(key, ".")

	for i := 0; i < len(kk)/2; i++ {
		s := kk[i]
		kk[i] = kk[len(kk)-i-1]
		kk[len(kk)-i-1] = s
	}
	for _, k := range kk {
		data = map[interface{}]interface{}{k: data}
	}
	return data
}

func InitManifest(logger lager.Logger, p Plan, instanceID string) error {
	os.Chmod(p.InitScriptPath, 755)
	cmd := exec.Command(p.InitScriptPath)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CREDENTIALS=secret/%s", instanceID))
	/* put more environment variables here, as needed */

	out, err := cmd.CombinedOutput()
	logger.Debug("initialized-manifest", lager.Data{
		"output": string(out),
	})
	return err
}

func GenManifest(p Plan, manifests ...map[interface{}]interface{}) (string, error) {
	var manifest map[interface{}]interface{}

	err := yaml.Unmarshal([]byte(p.RawManifest), &manifest)
	if err != nil {
		return "", err
	}

	merged, err := spruce.Merge(manifest)
	if err != nil {
		return "", err
	}
	for _, next := range manifests {
		merged, err = spruce.Merge(merged, next)
		if err != nil {
			return "", err
		}
	}
	eval := &spruce.Evaluator{Tree: merged}
	err = eval.Run([]string{})
	if err != nil {
		return "", err
	}
	final := eval.Tree

	b, err := yaml.Marshal(final)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
