package commands_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/heroku/color"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/spf13/cobra"

	h2 "github.com/buildpacks/imgutil/testhelpers"
	"github.com/buildpacks/pack/internal/commands"
	"github.com/buildpacks/pack/internal/commands/testmocks"
	"github.com/buildpacks/pack/internal/config"
	"github.com/buildpacks/pack/pkg/logging"
	h "github.com/buildpacks/pack/testhelpers"
)

const validLocalManifest = `
{
    "schemaVersion": 2,
    "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
    "manifests": [
        {
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "size": 528,
            "digest": "sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33",
            "platform": {
                "architecture": "amd64",
                "os": "linux"
            }
        },
        {
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "size": 528,
            "digest": "sha256:b312e4b0e2c665d634602411fcb7c2699ba748c36f59324457bc17de485f36f6",
            "platform": {
                "architecture": "arm64",
                "os": "linux"
            }
        }
    ]
}
`

var dockerRegistry, readonlyDockerRegistry, customRegistry *h2.DockerRegistry

func newTestIndexName(providedPrefix ...string) string {
	prefix := "pack-index-test"
	if len(providedPrefix) > 0 {
		prefix = providedPrefix[0]
	}

	return dockerRegistry.RepoName(prefix + "-" + h.RandString(10))
}

// Change a reference name string into a valid file name
// Ex: cnbs/sample-package:hello-multiarch-universe
// to cnbs_sample-package-hello-multiarch-universe
func makeFileSafeName(ref string) string {
	fileName := strings.ReplaceAll(ref, ":", "-")
	return strings.ReplaceAll(fileName, "/", "_")
}

func TestManifestCreateCommand(t *testing.T) {
	dockerConfigDir, err := ioutil.TempDir("", "test.docker.config.dir")
	h.AssertNil(t, err)
	defer os.RemoveAll(dockerConfigDir)

	sharedRegistryHandler := registry.New(registry.Logger(log.New(ioutil.Discard, "", log.Lshortfile)))
	dockerRegistry = h2.NewDockerRegistry(h2.WithAuth(dockerConfigDir), h2.WithSharedHandler(sharedRegistryHandler))

	dockerRegistry.SetInaccessible("cnbs/no-image-in-this-name")

	dockerRegistry.Start(t)
	defer dockerRegistry.Stop(t)

	os.Setenv("DOCKER_CONFIG", dockerRegistry.DockerDirectory)
	defer os.Unsetenv("DOCKER_CONFIG")

	color.Disable(true)
	defer color.Disable(false)
	spec.Run(t, "CreateCommand", testManifestCreateCommand, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testManifestCreateCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		command        *cobra.Command
		logger         logging.Logger
		outBuf         bytes.Buffer
		mockController *gomock.Controller
		mockClient     *testmocks.MockPackClient
		manifestDir    string
	)

	it.Before(func() {
		packHome, err := config.PackHome()
		h.AssertNil(t, err)

		manifestDir = filepath.Join(packHome, "manifests")

		mockController = gomock.NewController(t)
		mockClient = testmocks.NewMockPackClient(mockController)
		logger = logging.NewLogWithWriters(&outBuf, &outBuf)
		command = commands.ManifestCreate(logger, mockClient)
	})

	it.After(func() {
		mockController.Finish()
	})

	when("#Create", func() {
		when("--publish", func() {
			it("platform information is missing", func() {
				image := newTestIndexName()
				command.SetArgs([]string{
					image,
					"cnbs/sample-package:hello-universe",
					"cnbs/sample-package:hello-universe-windows",
					"--publish",
				})
				err := command.Execute()
				h.AssertError(t, err, "missing either OS or Architecture information")
			})

			it("succesfully publish to registry", func() {
				image := newTestIndexName()
				command.SetArgs([]string{
					image,
					"alpine@sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33",
					"alpine@sha256:b312e4b0e2c665d634602411fcb7c2699ba748c36f59324457bc17de485f36f6",
					"--publish",
				})
				err := command.Execute()
				h.AssertNil(t, err)
			})
		})

		when("no --publish flag", func() {
			it("create a local index", func() {
				image := newTestIndexName()
				command.SetArgs([]string{
					image,
					"alpine@sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33",
					"alpine@sha256:b312e4b0e2c665d634602411fcb7c2699ba748c36f59324457bc17de485f36f6",
				})
				err := command.Execute()
				h.AssertNil(t, err)

				testManifest := filepath.Join(manifestDir, makeFileSafeName(image))

				jsonFile, err := os.ReadFile(testManifest)
				h.AssertNil(t, err)
				h.AssertEq(t, jsonFile, []byte(validLocalManifest))
			})
		})

	})
}
