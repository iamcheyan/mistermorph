package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/clifmt"
	sharedcredits "github.com/quailyquaily/mistermorph/internal/credits"
	"github.com/spf13/cobra"
)

func newCreditsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "credits",
		Short: "Show open source and contributor credits",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := sharedcredits.Load()
			if err != nil {
				return err
			}
			printCredits(cmd.OutOrStdout(), data)
			return nil
		},
	}
}

func printCredits(out io.Writer, data sharedcredits.Data) {
	if out == nil {
		return
	}
	printOpenSourceCredits(out, data.OpenSource)
	if len(data.Contributors) > 0 {
		fmt.Fprintln(out)
		printContributorCredits(out, data.Contributors)
	}
}

func printOpenSourceCredits(out io.Writer, items []sharedcredits.OpenSourceEntry) {
	fmt.Fprintln(out, clifmt.Headerf("Open Source (%d)", len(items)))
	if len(items) == 0 {
		fmt.Fprintln(out, clifmt.Warn("No open source entries."))
		return
	}
	for _, item := range items {
		fmt.Fprintln(out, clifmt.Success(item.Name))
		fmt.Fprintf(out, "  License: %s\n", strings.TrimSpace(item.License))
		fmt.Fprintf(out, "  Link: %s\n", strings.TrimSpace(item.Link))
		fmt.Fprintf(out, "  %s\n\n", strings.TrimSpace(item.Summary))
	}
}

func printContributorCredits(out io.Writer, items []sharedcredits.ContributorEntry) {
	fmt.Fprintln(out, clifmt.Headerf("Contributors (%d)", len(items)))
	if len(items) == 0 {
		fmt.Fprintln(out, clifmt.Warn("No contributor entries."))
		return
	}
	for _, item := range items {
		fmt.Fprintln(out, clifmt.Success(item.Name))
		fmt.Fprintf(out, "  Link: %s\n", strings.TrimSpace(item.Link))
		fmt.Fprintf(out, "  %s\n\n", strings.TrimSpace(item.Summary))
	}
}
