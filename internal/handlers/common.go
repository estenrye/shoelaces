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
	"html/template"
	"net"
	"net/http"
	"strings"

	"github.com/thousandeyes/shoelaces/internal/mappings"

	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/ipxe"
	"github.com/thousandeyes/shoelaces/internal/utils"
)

// DefaultTemplateRenderer holds information for rendering a template based
// on its name. It implements the http.Handler interface.
type DefaultTemplateRenderer struct {
	templateName string
}

// RenderDefaultTemplate renders a template by the given name
func RenderDefaultTemplate(name string) *DefaultTemplateRenderer {
	return &DefaultTemplateRenderer{templateName: name}
}

func (t *DefaultTemplateRenderer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)
	tpl := env.StaticTemplates
	// XXX: Probably not ideal as it's doing the directory listing on every request
	ipxeScripts := ipxe.ScriptList(env)
	tplVars := struct {
		BaseURL      string
		HostnameMaps *[]mappings.HostnameMap
		NetworkMaps  *[]mappings.NetworkMap
		Scripts      *[]ipxe.Script
	}{
		env.BaseURL,
		&env.HostnameMaps,
		&env.NetworkMaps,
		&ipxeScripts,
	}
	renderTemplate(w, tpl, "header", tplVars)
	renderTemplate(w, tpl, t.templateName, tplVars)
	renderTemplate(w, tpl, "footer", tplVars)
}

func renderTemplate(w http.ResponseWriter, tpl *template.Template, tmpl string, d interface{}) {
	err := tpl.ExecuteTemplate(w, tmpl, d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func envFromRequest(r *http.Request) *environment.Environment {
	return r.Context().Value(ShoelacesEnvCtxID).(*environment.Environment)
}

func envNameFromRequest(r *http.Request) string {
	e := r.Context().Value(ShoelacesEnvNameCtxID)
	if e != nil {
		return e.(string)
	}
	return ""
}

func varMapFromRequest(r *http.Request) map[string]interface{} {
	variablesMap := map[string]interface{}{}

	env := r.Context().Value(ShoelacesEnvCtxID).(*environment.Environment)

	for key, val := range r.URL.Query() {
		env.Logger.Debug("URL_Query_Variable", key, "Value", val[0])
		variablesMap[key] = val[0]
		key_splits := strings.Split(key, ".")
		if len(key_splits) > 1 {
			map_pointer := map[string]interface{}{}

			for i, k := range key_splits {
				if i == 0 {
					if !utils.KeyInMap(k, variablesMap, env.Logger) {
						variablesMap[k] = map_pointer
					} else {
						map_pointer = variablesMap[k].(map[string]interface{})
					}
				} else if i < len(key_splits)-1 {
					if !utils.KeyInMap(k, map_pointer, env.Logger) {
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
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		s, _ := json.Marshal(variablesMap)
		env.Logger.Debug("Map", s)
		return variablesMap
	}

	x_forwrded_for := r.Header.Get("X-Forwarded-For")
	if x_forwrded_for != "" {
		ips := strings.Split(x_forwrded_for, ", ")
		if len(ips) >= 1 {
			ip = ips[0]
		}
	}

	variablesMap["ip_address"] = ip

	host := r.FormValue("host")
	if host == "" {
		host = resolveHostname(env.Logger, ip, env.DnsAddr)
	}

	variablesMap["hostname"] = host

	// Find with reverse hostname matched with the hostname regexps
	if script, found := mappings.FindScriptForHostname(env.HostnameMaps, host); found {
		for k, v := range script.Params {
			variablesMap[k] = v
		}
	} else if script, found := mappings.FindScriptForNetwork(env.NetworkMaps, ip); found {
		for k, v := range script.Params {
			variablesMap[k] = v
		}
	}

	s, _ := json.Marshal(variablesMap)
	env.Logger.Debug("Map", s)
	return variablesMap
}
