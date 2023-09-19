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

package utils

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
	"github.com/thousandeyes/shoelaces/internal/log"
)

// Filter receives a slide of strings and a function that receives a string
// and returns a bool, and returns a slide that has only the strings that
// returned true when they were applied the received function.
func Filter(files []string, fn func(string) bool) []string {
	var ret []string
	for _, f := range files {
		if fn(f) {
			ret = append(ret, f)
		}
	}

	return ret
}

// StringInSlice receives a string and a slice of strings and returns true if it exists
// there.
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// KeyInMap checks wheter the received key exists in the received map.
func KeyInMap(key string, mapInput map[string]interface{}, logger log.Logger) bool {
	logger.Debug("component", "utils", "function", "KeyInMap", "key", key)
	_, found := mapInput[key]
	key_splits := strings.Split(key, ".")
	logger.Debug("component", "utils", "function", "KeyInMap", "key", key, "found", found, "nested_key", (len(key_splits) > 1))
	if !found && len(key_splits) > 1 {
		temp_var, found := mapInput[key_splits[0]].(map[interface{}]interface{})
		logger.Debug("component", "utils", "function", "KeyInMap", "key", key, "found", found, "nested_key", (len(key_splits) > 1), "k", key_splits[0], "i", 0, "length", len(key_splits))
		if found {
			for i, k := range key_splits[1:] {
				if i < len(key_splits)-2 {
					temp_var, found = temp_var[k].(map[interface{}]interface{})
					var iface interface{} = temp_var
					logger.Debug("component", "utils", "function", "KeyInMap", "key", key, "found", found, "nested_key", (len(key_splits) > 1), "k", k, "i", i+1, "temp_var", iface)
					if found {
						continue
					}
					break
				} else {
					for sub_key, value := range temp_var {
						if sub_key == k {
							found = true
						} else {
							found = false
						}
						logger.Debug("component", "utils", "function", "KeyInMap", "key", key, "found", found, "nested_key", (len(key_splits) > 1), "k", k, "i", i+1, "key", sub_key.(string), "value", value.(string))
						if found {
							return true
						}
					}
				}
			}
		}
	}
	return found
}

// MapToString provides a string representation of a map of strings.
func MapToString(mapInput map[string]interface{}) string {
	result := ""
	for k, v := range mapInput {
		if len(result) > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s:%v", k, v)
	}
	return result
}

// BaseURLforEnvName provides an environment-sensitive method for returning
// the BaseURL of the application.
func BaseURLforEnvName(baseURL, environment string) string {
	if environment != "" {
		return filepath.Join(baseURL, "env", environment)
	}
	return baseURL
}

// ResolveHostname receives an IP and returns the resolved PTR. It returns an
// empty string in case the DNS lookup fails.
func ResolveHostname(logger log.Logger, ip string, dnsServer string) (host string) {
	addr, err := dns.ReverseAddr(ip)
	if err != nil {
		logger.Error("component", "utils", "function", "ResolveHostname", "msg", "Error resolving hostname", "ip", ip, "err", err, "addr", addr)
		return ""
	}

	logger.Debug("component", "utils", "function", "ResolveHostname", "msg", "Querying DNS server for PTR record", "ip", ip, "addr", addr, "dnsServer", dnsServer)
	m := new(dns.Msg)
	m.SetQuestion(addr, dns.TypePTR)
	m.RecursionDesired = true

	c := new(dns.Client)
	in, rtt, err := c.Exchange(m, dnsServer)
	if err != nil || len(in.Answer) == 0 {
		logger.Error("component", "utils", "function", "ResolveHostname", "msg", "Error resolving hostname", "ip", ip, "err", err)
		return ""
	}
	logger.Debug("component", "utils", "function", "ResolveHostname", "msg", "Hostname resolved", "ip", ip, "host", in.Answer[0].(*dns.PTR).Ptr, "rtt", rtt)
	return in.Answer[0].(*dns.PTR).Ptr
}

// IsValidIP returns whether or not an IP is well-formed.
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidMAC returns whether or not a MAC address is well-formed.
func IsValidMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}

// MacColonToDash receives a mac address and replace its colons by dashes
func MacColonToDash(mac string) string {
	return strings.Replace(mac, ":", "-", -1)
}

// MacDashToColon receives a mac address and replace its dashes by colons
func MacDashToColon(mac string) string {
	return strings.Replace(mac, "-", ":", -1)
}
