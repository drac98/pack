package commands

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	pubcfg "github.com/buildpacks/pack/config"

	"github.com/buildpacks/pack"
	"github.com/buildpacks/pack/internal/config"
	"github.com/buildpacks/pack/internal/style"
	"github.com/buildpacks/pack/logging"
)

func Rebase(logger logging.Logger, cfg config.Config, client PackClient) *cobra.Command {
	var opts pack.RebaseOptions
	var policy string

	cmd := &cobra.Command{
		Use:     "rebase <image-name>",
		Args:    cobra.ExactArgs(1),
		Short:   "Rebase app image with latest run image",
		Example: "pack rebase buildpacksio/pack",
		Long: "Rebase allows you to quickly swap out the underlying OS layers (run image) of an app image generated by `pack build` " +
			"with a newer version of the run image, without re-building the application.",
		RunE: LogError(logger, func(cmd *cobra.Command, args []string) error {
			opts.RepoName = args[0]
			opts.AdditionalMirrors = GetMirrors(cfg)

			var err error
			opts.PullPolicy, err = pubcfg.ParsePullPolicy(policy)
			if err != nil {
				return errors.Wrapf(err, "parsing pull policy %s", policy)
			}

			if err := client.Rebase(cmd.Context(), opts); err != nil {
				return err
			}
			logger.Infof("Successfully rebased image %s", style.Symbol(opts.RepoName))
			return nil
		}),
	}

	cmd.Flags().BoolVar(&opts.Publish, "publish", false, "Publish to registry")
	cmd.Flags().StringVar(&opts.RunImage, "run-image", "", "Run image to use for rebasing")
	cmd.Flags().StringVar(&policy, "pull-policy", "", "Pull policy to use. Accepted values are always, never, and if-not-present. The default is always")

	AddHelpFlag(cmd, "rebase")
	return cmd
}
