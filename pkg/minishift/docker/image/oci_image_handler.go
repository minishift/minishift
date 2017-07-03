/*
Copyright (C) 2017 Red Hat, Inc.

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

package image

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"bytes"
	"encoding/json"
	"github.com/containers/image/copy"
	"github.com/containers/image/oci/layout"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports/alltransports"
	"github.com/containers/image/types"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/sshutil"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	"github.com/minishift/minishift/pkg/util/progressdots"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// OciImageHandler is an ImageHandler implementation using OCI format to maintain the local cache.
type OciImageHandler struct {
	driver               drivers.Driver
	dockerClientSettings *dockerClientConfig
}

type dockerClientConfig struct {
	DockerHost      string
	DockerCertPath  string
	DockerTLSVerify bool
}

type Index struct {
	Manifests Manifests `json:"manifests"`
}

type Manifests []Manifest

type Manifest struct {
	Annotations Annotations `json:"annotations"`
}

type Annotations struct {
	Name string `json:"org.opencontainers.image.ref.name"`
}

// NewOciImageHandler creates a new ImageHandler which stores cached images in OCI format.
// It takes a reference to a Driver in order to communicate with the VM and Docker and a map containing the environment settings for the Minishift Docker daemon.
func NewOciImageHandler(driver drivers.Driver, dockerEnv map[string]string) (*OciImageHandler, error) {
	settings, err := getDockerSettings(dockerEnv)
	if err != nil {
		return nil, err
	}
	return &OciImageHandler{driver: driver, dockerClientSettings: settings}, nil
}

// ImportImages imports cached images from the host into the Docker daemon of the VM.
func (handler *OciImageHandler) ImportImages(config *ImageCacheConfig) error {
	out := handler.getOutputWriter(config)
	fmt.Fprintln(out, "-- Importing any available cached images")

	policyContext, err := handler.getPolicyContext()
	if err != nil {
		return fmt.Errorf("Error creating security context: %s", err.Error())
	}

	availableImages, err := handler.getDockerImages()
	if err != nil {
		return err
	}

	for _, imageName := range config.CachedImages {
		if _, found := availableImages[imageName]; found {
			continue
		}

		err := handler.importImage(imageName, config, policyContext, out)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportImages exports the images specified as part of the ImageCacheConfig from the VM to the host.
func (handler *OciImageHandler) ExportImages(config *ImageCacheConfig) error {
	out := handler.getOutputWriter(config)

	policyContext, err := handler.getPolicyContext()
	if err != nil {
		return fmt.Errorf("Error creating security context: %s", err.Error())
	}

	for _, imageName := range config.CachedImages {
		if handler.IsImageCached(config, imageName) {
			fmt.Fprint(out, fmt.Sprintf("%s found in cache", imageName))
		} else {
			err = handler.exportImage(imageName, config, policyContext, out)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsImageCached returns true if the specified image is cached, false otherwise.
func (handler *OciImageHandler) IsImageCached(config *ImageCacheConfig, image string) bool {
	cachedImages := handler.getCachedImages(config)
	_, found := cachedImages[image]
	return found
}

// AreImagesCached returns true if all images specified in the config are cached, false otherwise.
func (handler *OciImageHandler) AreImagesCached(config *ImageCacheConfig) bool {
	cachedImages := handler.getCachedImages(config)

	for _, image := range config.CachedImages {
		if _, found := cachedImages[image]; !found {
			return false
		}
	}

	return true
}

func (handler *OciImageHandler) getCachedImages(config *ImageCacheConfig) map[string]bool {
	cachedImages := make(map[string]bool)

	index, err := handler.getIndex(config)
	if index == nil || err != nil {
		return cachedImages
	}

	for _, manifest := range index.Manifests {
		cachedImages[manifest.Annotations.Name] = true
	}

	return cachedImages
}

func (handler *OciImageHandler) getDockerImages() (map[string]bool, error) {
	dockerImages := make(map[string]bool)

	session, err := handler.createSSHSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	var buffer bytes.Buffer
	session.Stdout = &buffer
	err = session.Run(cmd)
	if err != nil {
		return nil, fmt.Errorf("Error running command '%s': %v", cmd, err)
	}

	if len(buffer.String()) > 0 {
		for _, image := range strings.Split(buffer.String(), "\n") {
			dockerImages[image] = true
		}
	}

	return dockerImages, nil
}

func (handler *OciImageHandler) pullImage(image string, w io.Writer) error {
	defer util.TimeTrack(time.Now(), w, true)
	fmt.Fprint(w, fmt.Sprintf("Pulling image %s ", image))

	session, err := handler.createSSHSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("docker pull %s", image)
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("Error running command '%s': %v", cmd, err)
	}

	return nil
}

func (handler *OciImageHandler) getIndex(config *ImageCacheConfig) (*Index, error) {
	indexPath := filepath.Join(config.HostCacheDir, "index.json")
	if !filehelper.Exists(indexPath) {
		return nil, nil
	}

	raw, err := ioutil.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var index Index
	err = json.Unmarshal(raw, &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func (handler *OciImageHandler) importImage(image string, config *ImageCacheConfig, policyContext *signature.PolicyContext, out io.Writer) error {
	srcRef, err := layout.NewReference(config.HostCacheDir, image)
	if err != nil {
		return fmt.Errorf("Invalid image source '%v': %v", srcRef, err)
	}

	destRef, err := alltransports.ParseImageName(fmt.Sprintf("docker-daemon:%s", image))
	if err != nil {
		return fmt.Errorf("Invalid image source '%s': %v", image, err)
	}

	fmt.Fprint(out, fmt.Sprintf("   Importing %s ", image))
	progressDots := progressdots.New()
	progressDots.SetWriter(out)
	progressDots.Start()
	defer handler.endProgress(progressDots, out, err)
	err = handler.copyImage(srcRef, destRef, policyContext)
	if err != nil {
		return err
	}

	return nil
}

func (handler *OciImageHandler) exportImage(image string, config *ImageCacheConfig, policyContext *signature.PolicyContext, out io.Writer) error {
	availableImages, err := handler.getDockerImages()
	if err != nil {
		return err
	}

	if _, found := availableImages[image]; !found {
		handler.pullImage(image, config.Out)
	}

	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker-daemon:%s", image))
	if err != nil {
		return fmt.Errorf("Invalid image source '%s': %v", image, err)
	}

	destRef, err := layout.NewReference(config.HostCacheDir, image)
	if err != nil {
		return fmt.Errorf("Invalid image destination '%v': %v", destRef, err)
	}

	fmt.Fprint(out, fmt.Sprintf("Exporting %s", image))
	progressDots := progressdots.New()
	progressDots.SetWriter(out)
	progressDots.Start()
	defer handler.endProgress(progressDots, out, err)
	err = handler.copyImage(srcRef, destRef, policyContext)
	if err != nil {
		return err
	}

	return nil
}

func (handler *OciImageHandler) copyImage(srcRef types.ImageReference, destRef types.ImageReference, policyContext *signature.PolicyContext) error {
	err := copy.Image(policyContext, destRef, srcRef, &copy.Options{
		RemoveSignatures: false,
		SignBy:           "",
		ReportWriter:     nil,
		SourceCtx:        handler.getSystemContext(),
		DestinationCtx:   handler.getSystemContext(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (handler *OciImageHandler) getOutputWriter(config *ImageCacheConfig) io.Writer {
	var w io.Writer
	if config.Out != nil {
		w = config.Out
	} else {
		w = os.Stdout
	}
	return w
}

func (handler *OciImageHandler) getSystemContext() *types.SystemContext {
	return &types.SystemContext{
		DockerDaemonHost:                  handler.dockerClientSettings.DockerHost,
		DockerDaemonCertPath:              handler.dockerClientSettings.DockerCertPath,
		DockerDaemonInsecureSkipTLSVerify: !handler.dockerClientSettings.DockerTLSVerify,
	}
}

// createSSHSession creates an interactive SSH session
func (handler *OciImageHandler) createSSHSession() (*ssh.Session, error) {
	sshClient, err := sshutil.NewSSHClient(handler.driver)
	if err != nil {
		return nil, err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (handler *OciImageHandler) getPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error creating security context: %s", err.Error()))
	}

	return policyContext, nil
}

func (handler *OciImageHandler) endProgress(progressDots *progressdots.ProgressDots, out io.Writer, err error) {
	progressDots.Stop()
	if err == nil {
		fmt.Fprint(out, " OK")
	}
	fmt.Fprint(out, "\n")
}

func getDockerSettings(dockerEnv map[string]string) (*dockerClientConfig, error) {
	settings := &dockerClientConfig{}

	if val, ok := dockerEnv["DOCKER_HOST"]; ok {
		settings.DockerHost = val
	} else {
		return nil, errors.New("The provided Docker environment settings are missing the DOCKER_HOST key.")
	}

	if val, ok := dockerEnv["DOCKER_CERT_PATH"]; ok {
		settings.DockerCertPath = val
	} else {
		return nil, errors.New("The provided Docker environment settings are missing the DOCKER_CERT_PATH key.")
	}

	if val, ok := dockerEnv["DOCKER_TLS_VERIFY"]; ok {
		verify, err := strconv.ParseBool(val)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Invalid value '%s' for DOCKER_TLS_VERIFY key.", val))
		}
		settings.DockerTLSVerify = verify
	} else {
		return nil, errors.New("The provided Docker environment settings are missing the DOCKER_TLS_VERIFY key.")
	}

	return settings, nil
}
