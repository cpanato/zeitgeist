/*
Copyright 2020 The Kubernetes Authors.

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

package dependency

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

func TestUnsupported(t *testing.T) {
	client, err := NewLocalClient()
	require.Nil(t, err)
	_, err = client.RemoteCheck("")
	require.True(t, errors.As(err, &UnsupportedError{}))
	_, err = client.RemoteExport("")
	require.True(t, errors.As(err, &UnsupportedError{}))
	_, err = client.Upgrade("", "")
	require.True(t, errors.As(err, &UnsupportedError{}))
}

func TestLocalSuccess(t *testing.T) {
	client, err := NewLocalClient()
	require.Nil(t, err)

	err = client.LocalCheck("../testdata/local.yaml", "../testdata")
	require.Nil(t, err)
}

func TestRemoteUnsupported(t *testing.T) {
	_, err := NewRemoteClient()
	require.True(t, errors.As(err, &UnsupportedError{}))
}

func TestBrokenFile(t *testing.T) {
	client, err := NewLocalClient()
	require.Nil(t, err)

	err = client.LocalCheck("../testdata/does-not-exist", "../testdata")
	require.NotNil(t, err)

	err = client.LocalCheck("../testdata/Dockerfile", "../testdata")
	require.NotNil(t, err)
}

func TestLocalOutOfSync(t *testing.T) {
	client, err := NewLocalClient()
	require.Nil(t, err)

	err = client.LocalCheck("../testdata/local-out-of-sync.yaml", "../testdata")
	require.NotNil(t, err)
}

func TestLocalInvalid(t *testing.T) {
	client, err := NewLocalClient()
	require.Nil(t, err)

	err = client.LocalCheck("../testdata/local-invalid.yaml", "../testdata")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "compiling regex")
}

func TestFileDoesntExist(t *testing.T) {
	client, err := NewLocalClient()
	require.Nil(t, err)

	err = client.LocalCheck("../testdata/local-no-file.yaml", "../testdata")
	require.NotNil(t, err)
}

func TestDeserialising(t *testing.T) {
	invalidYamls := []string{
		"a b c",
		"name:",
		"name: test",
		"version: 1.0.0",
	}

	for _, invalid := range invalidYamls {
		var d Dependency

		err := yaml.Unmarshal([]byte(invalid), &d)
		require.NotNil(t, err)
	}

	validYamls := []string{
		"name: test\nversion: 1.0.0",
		"name: test\nversion: 100",
	}

	for _, valid := range validYamls {
		var d Dependency

		err := yaml.Unmarshal([]byte(valid), &d)
		require.Nil(t, err)
	}
}

func TestSetVersion(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")

	err := os.WriteFile(testFile, []byte("APP1_VERSION: 0.0.1\nAPP2_VERSION: 0.0.1"), 0o644)
	require.Nil(t, err)

	err = os.WriteFile(filepath.Join(dir, "dependencies.yaml"), []byte(`
dependencies:
  - name: app1
    version: 0.0.1
    scheme: semver
    upstream:
      flavour: dummy
      url: example/example
    refPaths:
    - path: test.txt
      match: APP1_VERSION
  - name: app2
    version: 0.0.1
    scheme: semver
    refPaths:
    - path: test.txt
      match: APP2_VERSION
`), 0o644)
	require.Nil(t, err)

	client, err := NewLocalClient()
	require.Nil(t, err)
	err = client.SetVersion(filepath.Join(dir, "dependencies.yaml"), dir, "app1", "2.1.0")
	if err != nil {
		t.Fatalf("SetVersion failed: %v", err)
	}

	got, err := os.ReadFile(testFile)
	require.Nil(t, err)
	require.Equal(t, "APP1_VERSION: 2.1.0\nAPP2_VERSION: 0.0.1", string(got))
}
