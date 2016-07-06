package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
)

func main() {
	client := github.NewClient(nil)
	release, resp, err := client.Repositories.GetLatestRelease("openshift", "origin")
	if err != nil {
		fmt.Printf("Could not get latest OpenShift release: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	assetID := getOpenShiftServerAssetID(release)
	if assetID == 0 {
		fmt.Println("Could not get OpenShift release URL")
		os.Exit(1)
	}
	asset, url, err := client.Repositories.DownloadReleaseAsset("openshift", "origin", assetID)
	if err != nil {
		fmt.Printf("Could not download OpenShift release asset: %s\n", err)
		os.Exit(1)
	}
	if len(url) > 0 {
		httpResp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Could not download OpenShift release asset: %s\n", err)
			os.Exit(1)
		}
		asset = httpResp.Body
	}

	defer asset.Close()

	gzf, err := gzip.NewReader(asset)
	if err != nil {
		fmt.Printf("Could not ungzip OpenShift release asset: %s\n", err)
		os.Exit(1)
	}
	defer gzf.Close()
	tr := tar.NewReader(gzf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			fmt.Printf("Could not extract OpenShift release asset: %s\n", err)
			os.Exit(1)
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != "kube-apiserver" {
			continue
		}
		contents, err := ioutil.ReadAll(tr)
		if err != nil {
			fmt.Printf("Could not extract OpenShift release asset: %s\n", err)
			os.Exit(1)
		}
		err = ioutil.WriteFile("out/openshift", contents, os.ModePerm)
		if err != nil {
			fmt.Printf("Could not write OpenShift binary: %s\n", err)
			os.Exit(1)
		}
	}
}

func getOpenShiftServerAssetID(release *github.RepositoryRelease) int {
	for _, asset := range release.Assets {
		if strings.HasPrefix(*asset.Name, "openshift-origin-server") && strings.HasSuffix(*asset.Name, "linux-64bit.tar.gz") {
			return *asset.ID
		}
	}
	return 0
}
