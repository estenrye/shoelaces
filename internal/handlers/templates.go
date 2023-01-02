// Copyright 2018 ThousandEyes Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/thousandeyes/shoelaces/internal/utils"
)

// TemplateHandler handles templated config files
type TemplateHandler struct{}

// TemplateHandler is the dynamic configuration provider endpoint. It
// receives a key and maybe an environment.
func (t *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	variablesMap := map[string]interface{}{}
	configName := filepath.Clean(r.URL.Path)

	if configName == "" {
		http.Error(w, "No template name provided", http.StatusNotFound)
		return
	}

	env := envFromRequest(r)
	envName := envNameFromRequest(r)

	for key, val := range r.URL.Query() {
		env.Logger.Info("URL_Query_Variable", key, "Value", val[0])
		variablesMap[key] = val[0]
		key_splits := strings.Split(key, ".")
		if len(key_splits) > 1 {
			map_pointer := map[string]interface{}{}

			for i, k := range key_splits {
				if i == 1 {
					if !utils.KeyInMap(k, variablesMap) {
						variablesMap[k] = map_pointer
					} else {
						map_pointer = variablesMap[k].(map[string]interface{})
					}
				} else if i < len(key_splits)-1 {
					if !utils.KeyInMap(k, map_pointer) {
						temp := map[string]interface{}{}
						map_pointer[k] = temp
						map_pointer = temp
					} else {
						map_pointer = map_pointer[k].(map[string]interface{})
					}
				} else {
					map_pointer[k] = val[0]
				}
			}
		}
		s, _ := json.Marshal(variablesMap)
		env.Logger.Info("Map", s)
	}

	variablesMap["baseURL"] = utils.BaseURLforEnvName(env.BaseURL, envName)

	configString, err := env.Templates.RenderTemplate(env.Logger, configName, variablesMap, envName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		io.WriteString(w, configString)
	}
}

// TemplateHandler returns a TemplateHandler instance implementing http.Handler
func TemplateServer() *TemplateHandler {
	return &TemplateHandler{}
}

// GetTemplateParams receives a script name and returns the parameters
// required for completing that template.
func GetTemplateParams(w http.ResponseWriter, r *http.Request) {
	var vars []string
	env := envFromRequest(r)

	filterBlacklist := func(s string) bool {
		return !utils.StringInSlice(s, env.ParamsBlacklist)
	}

	script := r.URL.Query().Get("script")
	if script == "" {
		http.Error(w, "Required script parameter", http.StatusInternalServerError)
		return
	}

	envName := r.URL.Query().Get("environment")
	if envName == "" {
		envName = "default"
	}

	vars = utils.Filter(env.Templates.ListVariables(script, envName), filterBlacklist)

	marshaled, err := json.Marshal(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(marshaled)
}
