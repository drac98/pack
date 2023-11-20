package commands_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/heroku/color"
	"github.com/pkg/errors"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/spf13/cobra"

	"github.com/buildpacks/imgutil"
	h2 "github.com/buildpacks/imgutil/testhelpers"
	"github.com/buildpacks/pack/internal/commands"
	"github.com/buildpacks/pack/internal/commands/testmocks"
	"github.com/buildpacks/pack/internal/config"
	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/logging"
	h "github.com/buildpacks/pack/testhelpers"
)

var (
	validImage1 = "alpine@sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33"
	validImage2 = "alpine@sha256:b312e4b0e2c665d634602411fcb7c2699ba748c36f59324457bc17de485f36f6"
)

var dockerRegistry, readonlyDockerRegistry, customRegistry *h2.DockerRegistry

func newTestIndexName(providedPrefix ...string) string {
	prefix := "pack-index-test"
	if len(providedPrefix) > 0 {
		prefix = providedPrefix[0]
	}

	return dockerRegistry.RepoName(prefix + "-" + h.RandString(10))
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
		when("flags", func() {
			when("invalid flag", func() {
				image := newTestIndexName()
				it("--format", func() {
					command.SetArgs([]string{
						image,
						validImage1,
						validImage2,
						"--format",
						"aNewMediaType",
					})
					err := command.Execute()
					h.AssertError(t, err, "unsupported media type given for --format")
				})

				it("--registry URL", func() {
					command.SetArgs([]string{
						image,
						validImage1,
						validImage1,
						"--registry",
						"host.@/org/repo/im@!ge:v1",
					})
					err := command.Execute()
					h.AssertError(t, err, "invalid registry URL ")
				})
			})

			when("valid flag", func() {
				image := newTestIndexName()
				it("--format OCI", func() {
					opts := client.CreateManifestOptions{
						ManifestName: image,
						Manifests:    []string{validImage1, validImage2},
						MediaType:    imgutil.OCITypes,
						Publish:      false,
						ManifestDir:  manifestDir,
					}

					mockClient.EXPECT().CreateManifest(gomock.Any(), opts).Return(nil).AnyTimes()

					command.SetArgs([]string{
						image,
						validImage1,
						validImage2,
						"--format",
						"oci",
					})
					err := command.Execute()
					h.AssertNil(t, err)
				})

				it("--publish", func() {
					opts := client.CreateManifestOptions{
						ManifestName: image,
						Manifests:    []string{validImage1, validImage2},
						MediaType:    imgutil.DockerTypes,
						Publish:      true,
						ManifestDir:  manifestDir,
					}

					mockClient.EXPECT().CreateManifest(gomock.Any(), opts).Return(nil).AnyTimes()
					command.SetArgs([]string{
						image,
						validImage1,
						validImage2,
						"--publish",
					})
					err := command.Execute()
					h.AssertNil(t, err)
				})
			})

			when("defaults", func() {
				image := newTestIndexName()
				it("success", func() {
					opts := client.CreateManifestOptions{
						ManifestName: image,
						Manifests:    []string{validImage1, validImage2},
						MediaType:    imgutil.DockerTypes,
						Publish:      false,
						ManifestDir:  manifestDir,
					}

					mockClient.EXPECT().CreateManifest(gomock.Any(), opts).Return(nil).AnyTimes()

					command.SetArgs([]string{
						image,
						validImage1,
						validImage2,
					})
					err := command.Execute()
					h.AssertNil(t, err)
				})
			})
		})

		it("CreateManifest", func() {
			image := newTestIndexName()

			mockClient.EXPECT().CreateManifest(gomock.Any(), gomock.Any()).Return(errors.Errorf("some error")).AnyTimes()
			command.SetArgs([]string{
				image,
				validImage1,
				validImage1})
			err := command.Execute()
			h.AssertError(t, err, "some error")
		})
	})
}
