// Package forgia provides embedded filesystem access to vault templates
// and slash commands. These are compiled into the binary via go:embed,
// eliminating runtime dependency on the modules/ directory.
package forgia

import (
	"embed"
	"io/fs"
)

//go:embed all:modules/vault-template
var vaultTemplateEmbed embed.FS

//go:embed all:modules/claude-commands
var slashCommandEmbed embed.FS

// VaultTemplateFS returns the embedded vault template filesystem.
// Root is "modules/vault-template/" — callers see config.toml, constitution.md, etc.
func VaultTemplateFS() fs.FS {
	sub, err := fs.Sub(vaultTemplateEmbed, "modules/vault-template")
	if err != nil {
		panic("embedded vault templates missing: " + err.Error())
	}
	return sub
}

// SlashCommandFS returns the embedded slash command filesystem.
// Root is "modules/claude-commands/" — callers see fd-new.md, fd-review.md, etc.
func SlashCommandFS() fs.FS {
	sub, err := fs.Sub(slashCommandEmbed, "modules/claude-commands")
	if err != nil {
		panic("embedded slash commands missing: " + err.Error())
	}
	return sub
}

// ListSlashCommands returns the names of all embedded slash commands (without .md extension).
func ListSlashCommands() []string {
	cmdFS := SlashCommandFS()
	var names []string
	fs.WalkDir(cmdFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == "." {
			return nil
		}
		name := path
		if len(name) > 3 && name[len(name)-3:] == ".md" {
			name = name[:len(name)-3]
		}
		names = append(names, name)
		return nil
	})
	return names
}

// ReadSlashCommand returns the content of a slash command by name.
// The name should not include the .md extension.
func ReadSlashCommand(name string) ([]byte, error) {
	return fs.ReadFile(SlashCommandFS(), name+".md")
}
