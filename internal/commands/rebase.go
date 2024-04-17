package commands

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/image"

	"github.com/buildpacks/pack/internal/config"
	"github.com/buildpacks/pack/internal/style"
	"github.com/buildpacks/pack/pkg/logging"
)

func Rebase(logger logging.Logger, cfg config.Config, pack PackClient) *cobra.Command {
	var opts client.RebaseOptions
	var policy string

	cmd := &cobra.Command{
		Use:     "rebase <image-name>",
		Args:    cobra.ExactArgs(1),
		Short:   "Rebase app image with latest run image",
		Example: "pack rebase buildpacksio/pack",
		Long: "Rebase allows you to quickly swap out the underlying OS layers (run image) of an app image generated by `pack build` " +
			"with a newer version of the run image, without re-building the application.",
		RunE: logError(logger, func(cmd *cobra.Command, args []string) error {
			opts.RepoName = args[0]
			opts.AdditionalMirrors = getMirrors(cfg)

			var err error
			stringPolicy := policy
			if stringPolicy == "" {
				stringPolicy = cfg.PullPolicy
			}
			opts.PullPolicy, err = image.ParsePullPolicy(stringPolicy)
			if err != nil {
				return errors.Wrapf(err, "parsing pull policy %s", stringPolicy)
			}

			if err := pack.Rebase(cmd.Context(), opts); err != nil {
				return err
			}
			logger.Infof("Successfully rebased image %s", style.Symbol(opts.RepoName))
			return nil
		}),
	}

	cmd.Flags().BoolVar(&opts.Publish, "publish", false, "Publish the rebased application image directly to the container registry specified in <image-name>, instead of the daemon. The previous application image must also reside in the registry.")
	cmd.Flags().StringVar(&opts.RunImage, "run-image", "", "Run image to use for rebasing")
	cmd.Flags().StringVar(&policy, "pull-policy", "", "Pull policy to use. Accepted values are always, never, and if-not-present. The default is always")
	cmd.Flags().StringVar(&opts.PreviousImage, "previous-image", "", "Image to rebase. Set to a particular tag reference, digest reference, or (when performing a daemon build) image ID. Use this flag in combination with <image-name> to avoid replacing the original image.")
	cmd.Flags().StringVar(&opts.ReportDestinationDir, "report-output-dir", "", "Path to export build report.toml.\nOmitting the flag yield no report file.")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Perform rebase operation without target validation (only available for API >= 0.12)")

	AddHelpFlag(cmd, "rebase")
	return cmd
}
