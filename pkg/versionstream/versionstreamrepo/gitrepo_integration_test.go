// +build integration

package versionstreamrepo_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/jenkins-x/jx/pkg/versionstream/versionstreamrepo"
	"github.com/stretchr/testify/assert"

	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
)

const (
	RepoURL           = "https://github.com/jenkins-x/jenkins-x-versions"
	TagFromDefaultURL = "v1.0.114"
	FirstTag          = "v0.0.1"
	SecondTag         = "v0.0.2"
	BranchRef         = "master"
	HEAD              = "HEAD"
)

func TestCloneJXVersionsRepoWithDefaultURL(t *testing.T) {
	origJxHome := os.Getenv("JX_HOME")

	tmpJxHome, err := ioutil.TempDir("", "jx-test-"+t.Name())
	assert.NoError(t, err)

	err = os.Setenv("JX_HOME", tmpJxHome)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpJxHome)
		err = os.Setenv("JX_HOME", origJxHome)
	}()

	gitter := gits.NewGitCLI()
	dir, versionRef, err := versionstreamrepo.CloneJXVersionsRepo(
		"",
		TagFromDefaultURL,
		nil,
		gitter,
		true,
		false,
		nil,
		nil,
		nil,
	)

	// Get the latest tag so that we know the correct expected verion ref.
	tag, _, err := gitter.Describe(dir, false, TagFromDefaultURL, "", true)

	assert.NoError(t, err)
	assert.NotNil(t, dir)
	assert.NotNil(t, versionRef)
	assert.Equal(t, tag, versionRef)
}

func initializeTempGitRepo(gitter gits.Gitter) (string, string, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", "", err
	}

	err = gitter.Init(dir)
	if err != nil {
		return "", "", err
	}

	err = gitter.AddCommit(dir, "Initial Commit")
	if err != nil {
		return "", "", err
	}

	testFile, err := ioutil.TempFile(dir, "versionstreams-test-")
	if err != nil {
		return "", "", err
	}

	testFileContents := []byte("foo")
	_, err = testFile.Write(testFileContents)
	if err != nil {
		return "", "", err
	}

	err = gitter.Add(dir, ".")
	if err != nil {
		return "", "", err
	}
	err = gitter.AddCommit(dir, "Adding foo")
	if err != nil {
		return "", "", err
	}

	err = gitter.CreateTag(dir, FirstTag, "First Tag")
	if err != nil {
		return "", "", err
	}

	testFileContents = []byte("bar")
	_, err = testFile.Write(testFileContents)
	if err != nil {
		return "", "", err
	}

	err = gitter.AddCommit(dir, "Adding bar")
	if err != nil {
		return "", "", err
	}

	err = gitter.CreateTag(dir, SecondTag, "Second Tag")
	if err != nil {
		return "", "", err
	}

	testFileContents = []byte("baz")
	_, err = testFile.Write(testFileContents)
	if err != nil {
		return "", "", err
	}

	err = gitter.AddCommit(dir, "Adding baz")
	if err != nil {
		return "", "", err
	}

	return fmt.Sprint(dir), filepath.Base(testFile.Name()), nil
}

func TestCloneJXVersionsRepoWithTeamSettings(t *testing.T) {
	origJxHome := os.Getenv("JX_HOME")

	tmpJxHome, err := ioutil.TempDir("", "jx-test-"+t.Name())
	assert.NoError(t, err)

	err = os.Setenv("JX_HOME", tmpJxHome)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpJxHome)
		err = os.Setenv("JX_HOME", origJxHome)
	}()

	gitter := gits.NewGitCLI()
	gitDir, testFileName, err := initializeTempGitRepo(gitter)
	defer func() {
		err := os.RemoveAll(gitDir)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)
	settings := &v1.TeamSettings{
		VersionStreamURL: gitDir,
		VersionStreamRef: FirstTag,
	}
	dir, versionRef, err := versionstreamrepo.CloneJXVersionsRepo(
		"",
		"",
		settings,
		gitter,
		true,
		false,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	assert.NotNil(t, versionRef)
	assert.Equal(t, FirstTag, versionRef)

	err = gitter.Checkout(dir, versionRef)
	assert.NoError(t, err)

	actualFileContents, err := ioutil.ReadFile(filepath.Join(dir, testFileName))
	assert.NoError(t, err)
	assert.Equal(t, "foo", string(actualFileContents))
}

func TestCloneJXVersionsRepoWithATag(t *testing.T) {
	origJxHome := os.Getenv("JX_HOME")

	tmpJxHome, err := ioutil.TempDir("", "jx-test-"+t.Name())
	assert.NoError(t, err)

	err = os.Setenv("JX_HOME", tmpJxHome)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpJxHome)
		err = os.Setenv("JX_HOME", origJxHome)
	}()

	gitter := gits.NewGitCLI()
	gitDir, testFileName, err := initializeTempGitRepo(gitter)
	defer func() {
		err := os.RemoveAll(gitDir)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)
	dir, versionRef, err := versionstreamrepo.CloneJXVersionsRepo(
		gitDir,
		FirstTag,
		nil,
		gitter,
		true,
		false,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	assert.NotNil(t, versionRef)
	assert.Equal(t, FirstTag, versionRef)

	err = gitter.Checkout(dir, versionRef)
	assert.NoError(t, err)

	actualFileContents, err := ioutil.ReadFile(filepath.Join(dir, testFileName))
	assert.NoError(t, err)
	assert.Equal(t, "foo", string(actualFileContents))
}

func TestCloneJXVersionsRepoWithABranch(t *testing.T) {
	origJxHome := os.Getenv("JX_HOME")

	tmpJxHome, err := ioutil.TempDir("", "jx-test-"+t.Name())
	assert.NoError(t, err)

	err = os.Setenv("JX_HOME", tmpJxHome)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpJxHome)
		err = os.Setenv("JX_HOME", origJxHome)
	}()

	gitter := gits.NewGitCLI()
	gitDir, testFileName, err := initializeTempGitRepo(gitter)
	defer func() {
		err := os.RemoveAll(gitDir)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)
	dir, versionRef, err := versionstreamrepo.CloneJXVersionsRepo(
		gitDir,
		BranchRef,
		nil,
		gitter,
		true,
		false,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	assert.NotNil(t, versionRef)
	assert.Equal(t, BranchRef, versionRef)

	err = gitter.Checkout(dir, versionRef)
	assert.NoError(t, err)

	actualFileContents, err := ioutil.ReadFile(filepath.Join(dir, testFileName))
	assert.NoError(t, err)
	assert.Equal(t, "foobarbaz", string(actualFileContents))
}

func TestCloneJXVersionsRepoWithACommit(t *testing.T) {
	origJxHome := os.Getenv("JX_HOME")

	tmpJxHome, err := ioutil.TempDir("", "jx-test-"+t.Name())
	assert.NoError(t, err)

	err = os.Setenv("JX_HOME", tmpJxHome)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpJxHome)
		err = os.Setenv("JX_HOME", origJxHome)
	}()

	gitter := gits.NewGitCLI()
	gitDir, testFileName, err := initializeTempGitRepo(gitter)
	defer func() {
		err := os.RemoveAll(gitDir)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)

	headMinusOne, err := gitter.RevParse(gitDir, "HEAD~1")

	dir, versionRef, err := versionstreamrepo.CloneJXVersionsRepo(
		fmt.Sprintf("file://%s", gitDir),
		headMinusOne,
		nil,
		gitter,
		true,
		false,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	assert.NotNil(t, versionRef)
	assert.Equal(t, SecondTag, versionRef)

	err = gitter.Checkout(dir, versionRef)
	assert.NoError(t, err)

	actualFileContents, err := ioutil.ReadFile(filepath.Join(dir, testFileName))
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(actualFileContents))
}

func TestCloneJXVersionsRepoWithAnUntaggedCommit(t *testing.T) {
	origJxHome := os.Getenv("JX_HOME")

	tmpJxHome, err := ioutil.TempDir("", "jx-test-"+t.Name())
	assert.NoError(t, err)

	err = os.Setenv("JX_HOME", tmpJxHome)
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tmpJxHome)
		err = os.Setenv("JX_HOME", origJxHome)
	}()

	gitter := gits.NewGitCLI()
	gitDir, testFileName, err := initializeTempGitRepo(gitter)
	defer func() {
		err := os.RemoveAll(gitDir)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)

	head, err := gitter.RevParse(gitDir, "HEAD")

	dir, versionRef, err := versionstreamrepo.CloneJXVersionsRepo(
		fmt.Sprintf("file://%s", gitDir),
		head,
		nil,
		gitter,
		true,
		false,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	assert.NotNil(t, versionRef)
	assert.Equal(t, head, versionRef)

	err = gitter.Checkout(dir, versionRef)
	assert.NoError(t, err)

	actualFileContents, err := ioutil.ReadFile(filepath.Join(dir, testFileName))
	assert.NoError(t, err)
	assert.Equal(t, []byte("foobarbaz"), actualFileContents)
}
