package client

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/heroku/color"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	ifakes "github.com/buildpacks/pack/internal/fakes"
	"github.com/buildpacks/pack/pkg/logging"
	"github.com/buildpacks/pack/pkg/testmocks"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestAddManifest(t *testing.T) {
	color.Disable(true)
	defer color.Disable(false)

	spec.Run(t, "build", testAddManifest, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testAddManifest(t *testing.T, when spec.G, it spec.S) {
	var (
		mockController   *gomock.Controller
		mockIndexFactory *testmocks.MockIndexFactory
		fakeImageFetcher *ifakes.FakeImageFetcher
		out              bytes.Buffer
		logger           logging.Logger
		subject          *Client
		err              error
		tmpDir           string
	)

	it.Before(func() {
		fakeImageFetcher = ifakes.NewFakeImageFetcher()
		logger = logging.NewLogWithWriters(&out, &out, logging.WithVerbose())
		mockController = gomock.NewController(t)
		mockIndexFactory = testmocks.NewMockIndexFactory(mockController)

		tmpDir, err = os.MkdirTemp("", "add-manifest-test")
		h.AssertNil(t, err)
		os.Setenv("XDG_RUNTIME_DIR", tmpDir)

		subject, err = NewClient(
			WithLogger(logger),
			WithFetcher(fakeImageFetcher),
			WithIndexFactory(mockIndexFactory),
			WithExperimental(true),
			WithKeychain(authn.DefaultKeychain),
		)
		h.AssertSameInstance(t, mockIndexFactory, subject.indexFactory)
		h.AssertNil(t, err)

		// Create a remote image to be fetched when adding to the image index
		fakeImage := h.NewFakeWithRandomUnderlyingV1Image(t, nil)
		fakeImageFetcher.RemoteImages["index.docker.io/pack/image:latest"] = fakeImage
	})
	it.After(func() {
		mockController.Finish()
		h.AssertNil(t, os.RemoveAll(tmpDir))
	})

	when("#AddManifest", func() {
		when("index doesn't exists", func() {
			it.Before(func() {
				mockIndexFactory.EXPECT().LoadIndex(gomock.Any(), gomock.Any()).Return(nil, errors.New("index not found locally"))
			})

			it("should return an error", func() {
				err = subject.AddManifest(
					context.TODO(),
					ManifestAddOptions{
						IndexRepoName: "pack/none-existent-index",
						RepoName:      "pack/image",
					},
				)
				h.AssertError(t, err, "index not found locally")
			})
		})

		when("index exists", func() {
			var indexRepoName string
			it.Before(func() {
				indexRepoName = h.NewRandomIndexRepoName()
			})

			when("no errors on save", func() {
				it.Before(func() {
					idx := h.RandomCNBIndex(t, indexRepoName, 1, 2)
					mockIndexFactory.EXPECT().LoadIndex(gomock.Eq(indexRepoName), gomock.Any()).Return(idx, nil)
				})

				it("adds the given image", func() {
					err = subject.AddManifest(
						context.TODO(),
						ManifestAddOptions{
							IndexRepoName: indexRepoName,
							RepoName:      "pack/image",
						},
					)
					h.AssertNil(t, err)
					h.AssertContains(t, out.String(), "successfully added to index: 'pack/image'")
				})

				it("error when invalid manifest reference name is used", func() {
					err = subject.AddManifest(
						context.TODO(),
						ManifestAddOptions{
							IndexRepoName: indexRepoName,
							RepoName:      "pack@@image",
						},
					)
					h.AssertNotNil(t, err)
					h.AssertError(t, err, "is not a valid manifest reference")
				})

				it("error when manifest reference doesn't exist in a registry", func() {
					err = subject.AddManifest(
						context.TODO(),
						ManifestAddOptions{
							IndexRepoName: indexRepoName,
							RepoName:      "pack/image-not-found",
						},
					)
					h.AssertNotNil(t, err)
					h.AssertError(t, err, "does not exist in registry")
				})
			})

			when("errors on save", func() {
				it.Before(func() {
					cnbIdx := h.NewMockImageIndex(t, indexRepoName, 1, 2)
					cnbIdx.ErrorOnSave = true
					mockIndexFactory.
						EXPECT().
						LoadIndex(gomock.Eq(indexRepoName), gomock.Any()).
						Return(cnbIdx, nil).
						AnyTimes()
				})

				it("error when manifest couldn't be saved locally", func() {
					err = subject.AddManifest(
						context.TODO(),
						ManifestAddOptions{
							IndexRepoName: indexRepoName,
							RepoName:      "pack/image",
						},
					)
					h.AssertNotNil(t, err)
					h.AssertError(t, err, "could not be saved in the local storage")
				})
			})
		})
	})
}
