/*
Copyright (C) 2016 Red Hat, Inc.

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

package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

	units "github.com/docker/go-units"

	"github.com/minishift/minishift/pkg/minikube/constants"
	minishiftConstants "github.com/minishift/minishift/pkg/minishift/constants"
	"github.com/minishift/minishift/pkg/util"
	stringUtils "github.com/minishift/minishift/pkg/util/strings"
)

func IsValidDriver(string, driver string) error {
	for _, d := range constants.SupportedVMDrivers {
		if driver == d {
			return nil
		}
	}
	return fmt.Errorf("Driver '%s' is not supported", driver)
}

func isValidHumanSize(size string) (bool, error) {

	if _, err := units.FromHumanSize(size); err != nil {
		return false, err
	}
	return true, nil
}

func isValidMemorySize(size string) (bool, error) {

	if _, err := units.RAMInBytes(size); err != nil {
		return false, err
	}
	return true, nil
}

type sizeValidationFunc func(size string) (bool, error)

func isPositiveAndValidSize(sizeValidation sizeValidationFunc, name string, size string, errorMessage string) error {
	if err := IsPositive(name, stringUtils.GetSignedNumbers(size)); err != nil {
		return err
	}

	if valid, err := sizeValidation(size); !valid {
		return fmt.Errorf(errorMessage, err)
	}
	return nil
}

func IsValidDiskSize(name string, diskSize string) error {
	return isPositiveAndValidSize(isValidHumanSize, name, diskSize, "Disk size is not valid: %v")
}

func IsValidMemorySize(name string, memorySize string) error {
	return isPositiveAndValidSize(isValidHumanSize, name, memorySize, "Memory size is not valid: %v")
}

func IsPositive(name string, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("%s:%v", name, err)
	}
	if i <= 0 {
		return fmt.Errorf("%s must be > 0", name)
	}
	return nil
}

func IsValidCIDR(name string, cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("Error parsing CIDR: %v", err)
	}
	return nil
}

func IsValidPath(name string, path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s path is not valid: %v", name, err)
	}
	return nil
}

func IsValidProxy(name string, uri string) error {
	if err := util.ValidateProxyURL(uri); err != nil {
		return fmt.Errorf("'%s' path is not valid: %v", name, err)
	}
	return nil
}

func IsValidUrl(_ string, isoURL string) error {
	if isoURL == minishiftConstants.B2dIsoAlias || isoURL == minishiftConstants.CentOsIsoAlias {
		return nil
	}
	_, err := url.ParseRequestURI(isoURL)
	if err != nil {
		return fmt.Errorf("'%s' url is not valid: %v", isoURL, err)
	}
	return nil
}

func IsValidIPv4Address(name string, address string) error {
	ip := net.ParseIP(address).To4()
	if ip == nil {
		return fmt.Errorf("%s IPv4 address is not valid: '%s'", name, address)
	}

	return nil
}

func IsValidNetmask(name string, mask string) error {
	err := fmt.Errorf("%s netmask is not valid: '%s'", name, mask)

	if stringUtils.HasOnlyNumbers(mask) {
		prefixSize, _ := strconv.Atoi(mask)
		if prefixSize == 0 || prefixSize > 25 {
			return err
		}
		return nil
	}

	// got something like '255.255.255.0' instead
	ip := net.ParseIP(mask).To4()
	if ip == nil {
		return err
	}
	prefixSize, _ := net.IPMask(ip).Size()
	if prefixSize == 0 {
		return err
	}

	return nil
}
