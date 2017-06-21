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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/minishift/minishift/pkg/minikube/sshutil"
	"github.com/minishift/minishift/pkg/util"
	"github.com/minishift/minishift/pkg/util/filehelper"
	minishiftStrings "github.com/minishift/minishift/pkg/util/strings"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// Repository name seperator for container images
	RepositorySeparator = '/'

	// Tag separator for container images
	TagSeparator = ':'

	// Used to separate image name parts in a single file name
	FileSeparator = "@"
)

// ImageHandler is responsible for the import and export of images into the Docker daemon of the VM
type ImageHandler interface {
	// Imports cached images from the host into the Docker daemon of the VM.
	ImportImages(config *ImageCacheConfig) error

	// Exports the images specified as part of the ImageCacheConfig from the VM to the host.
	ExportImages(config *ImageCacheConfig) error

	// AreImagesCached returns true if all images specified in the config are cached, false otherwise.
	AreImagesCached(config *ImageCacheConfig) bool
}

type DockerImageHandler struct {
	driver drivers.Driver
}

func NewDockerImageHandler(driver drivers.Driver) (*DockerImageHandler, error) {
	return &DockerImageHandler{driver: driver}, nil
}

func (handler *DockerImageHandler) ImportImages(config *ImageCacheConfig) error {
	out := handler.getOutputWriter(config)

	files, _ := ioutil.ReadDir(config.HostCacheDir)
	if len(files) > 0 {
		fmt.Fprintln(out, "-- Importing cached images ...")
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tmp") {
			continue
		}

		if !minishiftStrings.Contains(config.CachedImages, handler.fileNameToImageName(f.Name())) {
			continue
		}

		err := handler.importImage(filepath.Join(config.HostCacheDir, f.Name()), out)
		if err != nil {
			return err
		}
	}
	return nil
}

func (handler *DockerImageHandler) ExportImages(config *ImageCacheConfig) error {
	out := handler.getOutputWriter(config)
	for _, exportImage := range config.CachedImages {
		pulled, err := handler.IsPulled(exportImage)
		if err != nil {
			return err
		}

		if !pulled {
			if config.ImageMissStrategy == PULL {
				err := handler.pullImage(exportImage, config.Out)
				if err != nil {
					return err
				}
			} else {
				continue
			}
		}

		err = handler.exportImage(exportImage, config.HostCacheDir, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (handler *DockerImageHandler) AreImagesCached(config *ImageCacheConfig) bool {
	for _, image := range config.CachedImages {
		imageCacheFile := filepath.Join(config.HostCacheDir, handler.imageNameToFileName(image))
		if !handler.isCached(imageCacheFile) {
			return false
		}

	}

	return true
}

// IsPulled returns true is the specified image is already cached in the Docker daemon, false otherwise.
func (handler *DockerImageHandler) IsPulled(image string) (bool, error) {
	session, err := handler.createSshSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	cmd := fmt.Sprintf("docker images -q %s", image)
	var buffer bytes.Buffer
	session.Stdout = &buffer
	err = session.Run(cmd)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error running command '%s': %v", cmd, err))
	}

	if len(buffer.String()) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (handler *DockerImageHandler) getOutputWriter(config *ImageCacheConfig) io.Writer {
	var w io.Writer
	if config.Out != nil {
		w = config.Out
	} else {
		w = os.Stdout
	}
	return w
}

func (handler *DockerImageHandler) importImage(imagePath string, out io.Writer) error {
	defer util.TimeTrack(time.Now(), out)
	fmt.Fprint(out, fmt.Sprintf("   Importing %s ", handler.fileNameToImageName(imagePath)))

	session, err := handler.createSshSession()
	if err != nil {
		return err
	}

	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	errorCh := make(chan error)
	go func() {
		file, err := os.Open(imagePath)
		if err != nil {
			errorCh <- errors.New(fmt.Sprintf("Unable to open image: %v", err))
		}

		reader := bufio.NewReader(file)
		io.Copy(w, reader)
		w.Close()

		errorCh <- nil
	}()

	cmd := "docker load"
	err = session.Run(cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("Error running command '%s': %v", cmd, err))
	}

	err = <-errorCh
	if err != nil {
		return err
	}

	return nil
}

func (handler *DockerImageHandler) exportImage(image string, cacheDir string, w io.Writer) error {
	defer util.TimeTrack(time.Now(), w)
	fmt.Fprint(w, fmt.Sprintf("Exporting image %s ", image))

	imageCacheFile := filepath.Join(cacheDir, handler.imageNameToFileName(image))
	if handler.isCached(imageCacheFile) {
		fmt.Fprint(w, " [already cached] ")
		return nil
	}

	session, err := handler.createSshSession()
	if err != nil {
		return err
	}
	defer session.Close()

	r, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	errorCh := make(chan error)
	go func() {
		file, err := os.Create(imageCacheFile + ".tmp")
		if err != nil {
			errorCh <- errors.New(fmt.Sprintf("Unable to create temporary image file: %v", err))
		}

		w := bufio.NewWriter(file)
		io.Copy(w, r)

		w.Flush()
		file.Close()

		err = os.Rename(file.Name(), imageCacheFile)
		if err != nil {
			errorCh <- errors.New(fmt.Sprintf("Unable to create image file: %v", err))
		}

		errorCh <- nil
	}()

	cmd := fmt.Sprintf("docker save %s", image)
	err = session.Run(cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("Error running command '%s': %v", cmd, err))
	}

	err = <-errorCh
	if err != nil {
		return err
	}

	return nil
}

func (handler *DockerImageHandler) pullImage(image string, w io.Writer) error {
	defer util.TimeTrack(time.Now(), w)
	fmt.Fprint(w, fmt.Sprintf("Pulling image %s ", image))

	session, err := handler.createSshSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("docker pull %s", image)
	err = session.Run(cmd)
	if err != nil {
		return errors.New(fmt.Sprintf("Error running command '%s': %v", cmd, err))
	}

	return nil
}

// availableImages returns a slice of available images in the current Docker instance
func (handler *DockerImageHandler) availableImages() ([]string, error) {
	session, err := handler.createSshSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cmd := "docker images --format '{{.Repository}}:{{.Tag}}'"
	var buffer bytes.Buffer
	session.Stdout = &buffer
	err = session.Run(cmd)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error running command '%s': %v", cmd, err))
	}

	availableImages := strings.Split(buffer.String(), "\n")
	return availableImages, nil
}

// imageNameToFileName maps an image name in the format repository/name:tag to a file system name.
func (handler *DockerImageHandler) imageNameToFileName(image string) string {
	imageNameParts := strings.FieldsFunc(image, func(r rune) bool {
		switch r {
		case RepositorySeparator, TagSeparator:
			return true
		}
		return false
	})

	return strings.Join(imageNameParts, FileSeparator)
}

// isImageFile returns true if the denoted path represents a image file (based on file name).
func (handler *DockerImageHandler) isImageFile(path string) bool {
	fileName := filepath.Base(path)
	match, err := regexp.Match("^.*@.*@.*$", []byte(fileName))
	if err != nil {
		return false
	}
	return match
}

// fileNameToImageName maps a filename on the host to an image name in the format repository/name:tag.
func (handler *DockerImageHandler) fileNameToImageName(path string) string {
	fileName := filepath.Base(path)
	parts := strings.Split(fileName, FileSeparator)

	return parts[0] + string(RepositorySeparator) + parts[1] + string(TagSeparator) + parts[2]
}

// isCached returns true if the specified image (file) exists on the host.
func (handler *DockerImageHandler) isCached(imageCacheFile string) bool {
	return filehelper.Exists(imageCacheFile)
}

// createSshSession creates an interactive SSH session
func (handler *DockerImageHandler) createSshSession() (*ssh.Session, error) {
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
