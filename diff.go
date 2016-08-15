package main

import (
	"fmt"

	"github.com/skeema/tengo"
)

func init() {
	long := `Compares the schemas on database instance(s) to the corresponding
filesystem representation of them. The output is a series of DDL commands that,
if run on the instance, would cause the instances' schemas to now match the
ones in the filesystem.`

	Commands["diff"] = &Command{
		Name:    "diff",
		Short:   "Compare a DB instance's schemas and tables to the filesystem",
		Long:    long,
		Handler: DiffCommand,
	}
}

func DiffCommand(cfg *Config) error {
	return diff(cfg, make(map[string]bool))
}

func diff(cfg *Config, seen map[string]bool) error {
	if cfg.Dir.IsLeaf() {
		if err := cfg.PopulateTemporarySchema(); err != nil {
			return err
		}

		mods := tengo.StatementModifiers{
			NextAutoInc: tengo.NextAutoIncIfIncreased,
		}

		for _, t := range cfg.Targets() {
			if canConnect, err := t.CanConnect(); !canConnect {
				// TODO: option to ignore/skip erroring hosts instead of failing entirely
				return fmt.Errorf("Cannot connect to %s: %s", t.Instance, err)
			}

			for _, schemaName := range t.SchemaNames {
				fmt.Printf("-- Diff of %s %s vs %s/*.sql\n", t.Instance, schemaName, cfg.Dir)
				from, err := t.Schema(schemaName)
				if err != nil {
					return err
				}
				to, err := t.TemporarySchema()
				if err != nil {
					return err
				}
				diff := tengo.NewSchemaDiff(from, to)
				if from == nil {
					// We have to create a new Schema to emit a create statement for the
					// correct DB name. We can't use to.CreateStatement() because that would
					// emit a statement referring to _skeema_tmp!
					// TODO: support db options
					newFrom := &tengo.Schema{Name: schemaName}
					fmt.Printf("%s;\n", newFrom.CreateStatement())
				}
				for _, tableDiff := range diff.TableDiffs {
					stmt := tableDiff.Statement(mods)
					if stmt != "" {
						fmt.Printf("%s;\n", stmt)
					}
				}
				fmt.Println()
			}
		}

		if err := cfg.DropTemporarySchema(); err != nil {
			return err
		}
	} else {
		// Recurse into subdirs, avoiding duplication due to symlinks
		seen[cfg.Dir.Path] = true
		subdirs, err := cfg.Dir.Subdirs()
		if err != nil {
			return err
		}
		for n := range subdirs {
			subdir := subdirs[n]
			if !seen[subdir.Path] {
				if err := cfg.ChangeDir(&subdir); err != nil {
					return err
				}
				if err := diff(cfg, seen); err != nil {
					return err
				}
			}
		}
	}

	return nil
}