package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit()
	case "add-feed":
		runAddFeed()
	case "add-all":
		runAddAll()
	case "remove-feed":
		runRemoveFeed()
	case "list-feeds":
		runListFeeds()
	case "status":
		runStatus()
	case "update":
		runUpdate()
	case "fetch":
		runFetch()
	case "generate":
		runGenerate()
	case "prune":
		runPrune()
	case "verify":
		runVerify()
	case "import-opml":
		runImportOPML()
	case "export-opml":
		runExportOPML()
	case "version":
		fmt.Printf("rp version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Rogue Planet - Modern feed aggregator

Usage:
  rp <command> [flags]

Commands:
  init [-f FILE]    Initialize a new planet in the current directory
  add-feed <url>    Add a feed to the planet
  add-all -f FILE   Add multiple feeds from a file
  remove-feed <url> Remove a feed from the planet
  list-feeds        List all configured feeds
  status            Show planet status (feed and entry counts)
  update            Fetch all feeds and regenerate site
  fetch             Fetch all feeds without generating
  generate          Generate site without fetching
  prune             Remove old entries from database
  verify            Validate configuration and environment
  import-opml FILE  Import feeds from OPML file
  export-opml       Export feeds to OPML format
  version           Show version information
  help              Show this help message

Init Flags:
  -f FILE           Import feeds from file (one URL per line)

Add-All Flags:
  -f FILE           Path to feeds file (one URL per line)

Import-OPML Flags:
  --dry-run         Preview feeds without importing

Export-OPML Flags:
  --output FILE     Output file (default: stdout)

Global Flags:
  --config <path>   Path to config file (default: ./config.ini)
  --verbose         Enable verbose logging
  --quiet           Only show errors

Examples:
  rp init
  rp init -f feeds.txt
  rp add-feed https://blog.golang.org/feed.atom
  rp add-all -f feeds.txt
  rp list-feeds
  rp status
  rp update
  rp generate --days 14
  rp prune --days 90
  rp import-opml feeds.opml
  rp import-opml feeds.opml --dry-run
  rp export-opml --output feeds.opml

`)
}

func runInit() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	feedsFile := fs.String("f", "", "Import feeds from file")
	fs.Parse(os.Args[2:])

	opts := InitOptions{
		FeedsFile:  *feedsFile,
		ConfigPath: "config.ini",
		Output:     os.Stdout,
	}

	if err := cmdInit(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runAddFeed() {
	fs := flag.NewFlagSet("add-feed", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: rp add-feed <url>")
		os.Exit(1)
	}

	opts := AddFeedOptions{
		URL:        fs.Arg(0),
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	if err := cmdAddFeed(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runAddAll() {
	fs := flag.NewFlagSet("add-all", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	feedsFile := fs.String("f", "", "Path to feeds file")
	fs.Parse(os.Args[2:])

	if *feedsFile == "" {
		fmt.Fprintln(os.Stderr, "Usage: rp add-all -f <feeds-file>")
		os.Exit(1)
	}

	opts := AddAllOptions{
		FeedsFile:  *feedsFile,
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	if err := cmdAddAll(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runRemoveFeed() {
	fs := flag.NewFlagSet("remove-feed", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: rp remove-feed <url>")
		os.Exit(1)
	}

	opts := RemoveFeedOptions{
		URL:        fs.Arg(0),
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	if err := cmdRemoveFeed(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runListFeeds() {
	fs := flag.NewFlagSet("list-feeds", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	fs.Parse(os.Args[2:])

	opts := ListFeedsOptions{
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	if err := cmdListFeeds(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runStatus() {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	fs.Parse(os.Args[2:])

	opts := StatusOptions{
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	if err := cmdStatus(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runUpdate() {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	fs.Parse(os.Args[2:])

	opts := UpdateOptions{
		ConfigPath: *configPath,
		Verbose:    *verbose,
		Output:     os.Stdout,
	}

	if err := cmdUpdate(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runFetch() {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	fs.Parse(os.Args[2:])

	opts := FetchOptions{
		ConfigPath: *configPath,
		Verbose:    *verbose,
		Output:     os.Stdout,
	}

	if err := cmdFetch(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runGenerate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	days := fs.Int("days", 0, "Number of days to include (overrides config)")
	fs.Parse(os.Args[2:])

	opts := GenerateOptions{
		ConfigPath: *configPath,
		Days:       *days,
		Output:     os.Stdout,
	}

	if err := cmdGenerate(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runPrune() {
	fs := flag.NewFlagSet("prune", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	days := fs.Int("days", 90, "Remove entries older than N days")
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without deleting")
	fs.Parse(os.Args[2:])

	opts := PruneOptions{
		ConfigPath: *configPath,
		Days:       *days,
		DryRun:     *dryRun,
		Output:     os.Stdout,
	}

	if err := cmdPrune(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	fs.Parse(os.Args[2:])

	opts := VerifyOptions{
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	if err := cmdVerify(opts); err != nil {
		os.Exit(1)
	}
}

func runImportOPML() {
	fs := flag.NewFlagSet("import-opml", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	dryRun := fs.Bool("dry-run", false, "Preview feeds without importing")
	fs.Parse(os.Args[2:])

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: rp import-opml <opml-file> [--dry-run]")
		os.Exit(1)
	}

	opts := ImportOPMLOptions{
		OPMLFile:   fs.Arg(0),
		ConfigPath: *configPath,
		DryRun:     *dryRun,
		Output:     os.Stdout,
	}

	if err := cmdImportOPML(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func runExportOPML() {
	fs := flag.NewFlagSet("export-opml", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	output := fs.String("output", "", "Output file (default: stdout)")
	fs.Parse(os.Args[2:])

	opts := ExportOPMLOptions{
		ConfigPath: *configPath,
		OutputFile: *output,
		Output:     os.Stdout,
	}

	if err := cmdExportOPML(opts); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
