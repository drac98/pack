package commands

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/buildpacks/pack/internal/config"
	"github.com/buildpacks/pack/internal/style"
	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/logging"
)

type ManifestAnnotateFlags struct {
	Architecture string // Set the architecture
	OS           string // Set the operating system
	Variant      string // Set the architecture variant
}

func ManifestAnnotate(logger logging.Logger, pack PackClient) *cobra.Command {
	var flags ManifestAnnotateFlags
	cmd := &cobra.Command{
		Use:     "annotate [OPTIONS] <manifest-list> <manifest>",
		Short:   "Annotate a manifest list",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		Example: `pack manifest annotate cnbs/sample-package:hello-universe-multiarch \ cnbs/sample-package:hello-universe --arch amd64`,
		Long:    "manifest annotate modifies a manifest list (Image index) and update the platform information for an image included in the manifest list.",
		RunE: logError(logger, func(cmd *cobra.Command, args []string) error {
			if err := validateManifestAnnotateFlags(&flags); err != nil {
				return err
			}

			indexName := args[0]
			manifest := args[1]

			packHome, err := config.PackHome()
			if err != nil {
				return err
			}

			manifestDir := filepath.Join(packHome, "manifests")

			if err := pack.AnnotateManifest(cmd.Context(), client.AnnotateManifestOptions{
				Index:        indexName,
				Path:         manifestDir,
				Manifest:     manifest,
				Architecture: flags.Architecture,
				OS:           flags.OS,
				Variant:      flags.Variant,
			}); err != nil {
				return err
			}
			logger.Infof("Successfully annotated image index %s", style.Symbol(indexName))

			return nil
		}),
	}

	cmd.Flags().StringVar(&flags.Architecture, "arch", "", "Set the architecture")
	cmd.Flags().StringVar(&flags.OS, "os", "", "Set the operating system")
	cmd.Flags().StringVar(&flags.Variant, "variant", "", "Set the architecture variant")

	AddHelpFlag(cmd, "annotate")
	return cmd
}

func validateManifestAnnotateFlags(p *ManifestAnnotateFlags) error {
	return nil
}
