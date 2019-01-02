package commands

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"cloud-mta-build-tool/internal/buildops"
	"cloud-mta-build-tool/internal/fs"
	"cloud-mta-build-tool/mta"
)

// CommandList - list of command to execute
type CommandList struct {
	Info    string
	Command []string
}

// CommandProvider - Get build command's to execute
//noinspection GoExportedFuncWithUnexportedType
func CommandProvider(modules mta.Module) (CommandList, error) {
	// Get config from ./commands_cfg.yaml as generated artifacts from source
	commands, err := parse(CommandsConfig)
	if err != nil {
		return CommandList{}, errors.Wrap(err, "failed to parse the commands configuration file")
	}
	customCommands, err := parse(CustomCommandsConfig)
	if err != nil {
		return CommandList{}, errors.Wrap(err, "failed to parse the custom commands configuration file")
	}
	return mesh(modules, commands, customCommands)
}

// Match the object according to type and provide the respective command
func mesh(module mta.Module, commands Builders, customCommands Builders) (CommandList, error) {
	// The object support deep struct for future use, can be simplified to flat object
	var cmds CommandList
	builder, custom := buildops.GetBuilder(&module)

	var actualCommands Builders
	var configStr string
	if custom {
		actualCommands = customCommands
		configStr = "custom commands"
	} else {
		actualCommands = commands
		configStr = "commands"
	}
	for _, b := range actualCommands.Builders {
		// Return only matching types
		if builder == b.Name {
			cmds.Info = b.Info
			for _, cmd := range b.Type {
				cmds.Command = append(cmds.Command, cmd.Command)
			}
			return cmds, nil
		}
	}

	return cmds, fmt.Errorf("the %s builder is not defined in the %s configuration", builder, configStr)
}

// CmdConverter - path and commands to execute
func CmdConverter(mPath string, cmdList []string) [][]string {
	var cmd [][]string
	for i := 0; i < len(cmdList); i++ {
		cmd = append(cmd, append([]string{mPath}, strings.Split(cmdList[i], " ")...))
	}
	return cmd
}

// GetModuleAndCommands - Get module from mta.yaml and
// commands (with resolved paths) configured for the module type
func GetModuleAndCommands(loc dir.IMtaParser, module string) (*mta.Module, []string, error) {
	mtaObj, err := loc.ParseFile()
	if err != nil {
		return nil, nil, err
	}
	// Get module respective command's to execute
	return moduleCmd(mtaObj, module)
}

// Get commands for specific module type
func moduleCmd(mta *mta.MTA, moduleName string) (*mta.Module, []string, error) {
	for _, m := range mta.Modules {
		if m.Name == moduleName {
			commandProvider, err := CommandProvider(*m)
			if err != nil {
				return nil, nil, err
			}
			return m, commandProvider.Command, nil
		}
	}
	return nil, nil, errors.Errorf("the %v module is not defined in the .mta file", moduleName)
}
