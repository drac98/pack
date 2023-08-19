package commands_test

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/spf13/cobra"

	"github.com/buildpacks/pack/internal/commands"
	"github.com/buildpacks/pack/internal/commands/testmocks"
	"github.com/buildpacks/pack/pkg/logging"
	h "github.com/buildpacks/pack/testhelpers"
)

func TestManifestCommand(t *testing.T) {
	spec.Run(t, "ManifestCommand", testManifestCommand, spec.Parallel(), spec.Report(report.Terminal{}))
}

func testManifestCommand(t *testing.T, when spec.G, it spec.S) {
	var (
		cmd    *cobra.Command
		logger logging.Logger
		outBuf bytes.Buffer
	)

	it.Before(func() {
		logger = logging.NewLogWithWriters(&outBuf, &outBuf)
		mockController := gomock.NewController(t)
		mockClient := testmocks.NewMockPackClient(mockController)
		cmd = commands.NewManifestCommand(logger, mockClient)
		cmd.SetOut(logging.GetWriterForLevel(logger, logging.InfoLevel))
	})

	when("manifest", func() {
		it("prints help text", func() {
			cmd.SetArgs([]string{})
			h.AssertNil(t, cmd.Execute())
			output := outBuf.String()
			h.AssertContains(t, output, "Handle manifest list")
			h.AssertContains(t, output, "Usage:")
			for _, command := range []string{"create", "annotate", "add", "push", "rm", "remove", "inspect"} {
				h.AssertContains(t, output, command)
				h.AssertNotContains(t, output, command+"-manifest")
			}
		})
	})
}
