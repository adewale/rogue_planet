package main

import (
	"flag"
	"fmt"

	"github.com/adewale/rogue_planet/pkg/logging"
)

// Flag parsing functions - extracted for testability
// Each function takes args []string and returns (Options, error)

func parseInitFlags(args []string) (InitOptions, error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	feedsFile := fs.String("f", "", "Import feeds from file")

	if err := fs.Parse(args); err != nil {
		return InitOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return InitOptions{
		FeedsFile:  *feedsFile,
		ConfigPath: "config.ini",
	}, nil
}

func parseAddFeedFlags(args []string) (AddFeedOptions, error) {
	fs := flag.NewFlagSet("add-feed", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")

	if err := fs.Parse(args); err != nil {
		return AddFeedOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	if fs.NArg() < 1 {
		return AddFeedOptions{}, fmt.Errorf("missing feed URL argument")
	}

	return AddFeedOptions{
		URL:        fs.Arg(0),
		ConfigPath: *configPath,
	}, nil
}

func parseAddAllFlags(args []string) (AddAllOptions, error) {
	fs := flag.NewFlagSet("add-all", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	feedsFile := fs.String("f", "", "Path to feeds file")

	if err := fs.Parse(args); err != nil {
		return AddAllOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	if *feedsFile == "" {
		return AddAllOptions{}, fmt.Errorf("missing feeds file argument")
	}

	return AddAllOptions{
		FeedsFile:  *feedsFile,
		ConfigPath: *configPath,
	}, nil
}

func parseRemoveFeedFlags(args []string) (RemoveFeedOptions, error) {
	fs := flag.NewFlagSet("remove-feed", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	force := fs.Bool("force", false, "Skip confirmation prompt")

	if err := fs.Parse(args); err != nil {
		return RemoveFeedOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	if fs.NArg() < 1 {
		return RemoveFeedOptions{}, fmt.Errorf("missing feed URL argument")
	}

	return RemoveFeedOptions{
		URL:        fs.Arg(0),
		ConfigPath: *configPath,
		Force:      *force,
	}, nil
}

func parseListFeedsFlags(args []string) (ListFeedsOptions, error) {
	fs := flag.NewFlagSet("list-feeds", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")

	if err := fs.Parse(args); err != nil {
		return ListFeedsOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return ListFeedsOptions{
		ConfigPath: *configPath,
	}, nil
}

func parseStatusFlags(args []string) (StatusOptions, error) {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")

	if err := fs.Parse(args); err != nil {
		return StatusOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return StatusOptions{
		ConfigPath: *configPath,
	}, nil
}

func parseUpdateFlags(args []string) (UpdateOptions, error) {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	if err := fs.Parse(args); err != nil {
		return UpdateOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return UpdateOptions{
		ConfigPath: *configPath,
		Verbose:    *verbose,
		Logger:     logging.New("info"),
	}, nil
}

func parseFetchFlags(args []string) (FetchOptions, error) {
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")

	if err := fs.Parse(args); err != nil {
		return FetchOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return FetchOptions{
		ConfigPath: *configPath,
		Verbose:    *verbose,
		Logger:     logging.New("info"),
	}, nil
}

func parseGenerateFlags(args []string) (GenerateOptions, error) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	days := fs.Int("days", 0, "Number of days to include (overrides config)")

	if err := fs.Parse(args); err != nil {
		return GenerateOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return GenerateOptions{
		ConfigPath: *configPath,
		Days:       *days,
	}, nil
}

func parsePruneFlags(args []string) (PruneOptions, error) {
	fs := flag.NewFlagSet("prune", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	days := fs.Int("days", 90, "Remove entries older than N days")
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without deleting")

	if err := fs.Parse(args); err != nil {
		return PruneOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return PruneOptions{
		ConfigPath: *configPath,
		Days:       *days,
		DryRun:     *dryRun,
	}, nil
}

func parseVerifyFlags(args []string) (VerifyOptions, error) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")

	if err := fs.Parse(args); err != nil {
		return VerifyOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return VerifyOptions{
		ConfigPath: *configPath,
	}, nil
}

func parseImportOPMLFlags(args []string) (ImportOPMLOptions, error) {
	fs := flag.NewFlagSet("import-opml", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	dryRun := fs.Bool("dry-run", false, "Preview feeds without importing")

	if err := fs.Parse(args); err != nil {
		return ImportOPMLOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	if fs.NArg() < 1 {
		return ImportOPMLOptions{}, fmt.Errorf("missing OPML file argument")
	}

	return ImportOPMLOptions{
		OPMLFile:   fs.Arg(0),
		ConfigPath: *configPath,
		DryRun:     *dryRun,
	}, nil
}

func parseExportOPMLFlags(args []string) (ExportOPMLOptions, error) {
	fs := flag.NewFlagSet("export-opml", flag.ContinueOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	output := fs.String("output", "", "Output file (default: stdout)")

	if err := fs.Parse(args); err != nil {
		return ExportOPMLOptions{}, fmt.Errorf("parsing flags: %w", err)
	}

	return ExportOPMLOptions{
		ConfigPath: *configPath,
		OutputFile: *output,
	}, nil
}
