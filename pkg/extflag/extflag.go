package extflag

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// RegisterPathOrContent register two flags to get configuration from file or command line.
func RegisterPathOrContent(flagSet *pflag.FlagSet, flagName string, help string) {
	fileFlagName := fmt.Sprintf("%s-file", flagName)
	contentFlagName := flagName

	fileHelp := fmt.Sprintf("Path to %s", help)
	flagSet.String(fileFlagName, "", fileHelp)

	contentHelp := fmt.Sprintf("Alternative to '%s' flag (lower priority). Content of %s", fileFlagName, help)
	flagSet.String(contentFlagName, "", contentHelp)
}

// LoadContent can load content from registered flag.
func LoadContent(cmd *cobra.Command, flagName string, required bool) ([]byte, error) {
	contentFlagName := flagName
	fileFlagName := fmt.Sprintf("%s-file", flagName)
	path, err := cmd.Flags().GetString(fileFlagName)
	if err != nil {
		return nil, errors.Wrapf(err, "get flag %s error", fileFlagName)
	}
	contentStr, err := cmd.Flags().GetString(contentFlagName)
	if err != nil {
		return nil, errors.Wrapf(err, "get flag %s error", contentFlagName)
	}

	if len(path) > 0 && len(contentStr) > 0 {
		return nil, errors.Errorf("both %s and %s flags set.", fileFlagName, contentFlagName)
	}

	var content []byte
	if len(path) > 0 {
		c, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "loading YAML file %s for %s", path, fileFlagName)
		}
		content = c
	} else {
		content = []byte(contentStr)
	}

	if len(content) == 0 && required {
		return nil, errors.Errorf("flag %s or %s is required for running this command and content cannot be empty.", fileFlagName, contentFlagName)
	}

	return content, nil
}
