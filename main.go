package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func overrideCmd() *cobra.Command {
	var refresh bool
	var refreshInterval time.Duration

	rootCmd := &cobra.Command{
		Use:   "hosts-override [HOST_NAME,(IP|RESOLVABLE_HOST_NAME)...]",
		Short: "Override hosts file entries for the lifetime of the process",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			clearScreen()
			entries := parseArgs(&args)
			file := hostsFileLocation()
			createHostsBackup(file)
			removeOverrides(file) // Fixes unclean shutdown
			expandedEntries := parseOverrides(entries, false)
			appendOverrides(file, expandedEntries)
			displayStatus(&refresh, refreshInterval, expandedEntries)

			if refresh {
				refreshTicker := time.NewTicker(refreshInterval)

				go func() {
					for {
						select {
						case <-refreshTicker.C:
							expandedEntries := parseOverrides(entries, true)

							if expandedEntries != nil {
								removeOverrides(file)
								clearScreen()
								appendOverrides(file, expandedEntries)
								displayStatus(&refresh, refreshInterval, expandedEntries)
							}
						}
					}
				}()
			}

			waitUntilExit()
			removeOverrides(file)
		},
	}

	rootCmd.Flags().BoolVarP(&refresh, "refresh", "r", false, "Refresh unresolved hosts")
	rootCmd.Flags().DurationVarP(&refreshInterval, "refresh-interval", "i", time.Duration(5)*time.Minute, "Refresh Interval")

	return rootCmd
}

func main() {
	overrideCmd().Execute()
}

func parseArgs(args *[]string) *hostsFileEntries {
	var entries hostsFileEntries

	for _, pair := range *args {
		hv := strings.Split(pair, ",")
		entries = append(entries, &hostsFileEntry{&hv[0], &hv[1], nil})
	}

	return &entries
}

func parseOverrides(entries *hostsFileEntries, continueOnError bool) *hostsFileEntries {
	expandedEntries := hostsFileEntries{}

	for _, entry := range *entries {
		if maybeIP := *maybeIP(entry.ip); maybeIP != "" {
			expandedEntries = append(expandedEntries, entry)
		} else {
			// NOTE: Try to resolve google.com, but this could be any domain to test
			//       if there is an Internet connection and hosts are resolvable.
			//
			//       Can't use user provided domains, as they can already be in the
			//       hosts file and will provide a false positive.
			_, googleErr := net.LookupIP("google.com")
			ips, err := net.LookupIP(*entry.ip)

			if googleErr != nil || err != nil {
				if continueOnError {
					return nil
				}

				fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
				os.Exit(1)
			}

			for _, ip := range ips {
				ip := ip.String()
				expandedEntries = append(
					expandedEntries,
					&hostsFileEntry{hostname: entry.hostname, ip: &ip, ipResovledFrom: entry.ip},
				)
			}
		}
	}

	return &expandedEntries
}

func appendOverrides(hostsFileLocation *string, entries *hostsFileEntries) {
	f, err := os.OpenFile(*hostsFileLocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err)
	}

	defer f.Close()

	if _, err := f.WriteString(*entriesAsString(entries)); err != nil {
		log.Println(err)
	}
}

func removeOverrides(hostsFileLocation *string) {
	contents, err := ioutil.ReadFile(*hostsFileLocation)

	if err != nil {
		fmt.Println(err)
		return
	}

	re := regexp.MustCompile("(?s)(" + startComment() + ").*(" + finishComment() + ")")
	removedOverrides := re.ReplaceAll(contents, []byte(""))

	if err := ioutil.WriteFile(*hostsFileLocation, removedOverrides, 0); err != nil {
		log.Println(err)
	}
}

func displayStatus(refresh *bool, refreshInterval time.Duration, entries *hostsFileEntries) {
	fmt.Println("\nhosts-override: Overriding hosts file entries for the lifetime of the process")

	if *refresh == true {
		fmt.Println("\n(Refreshing every " + refreshInterval.String() + ")...")
	}
	fmt.Println("\n" + *entriesAsString(entries) + "\n")
	fmt.Println("\nPress CTRL-C to exit gracefully (hosts file will reset)")
}

func wrappingComment(custom string) string {
	return "\n#########################\n" +
		"# hosts-override " + custom +
		"\n#########################\n\n"
}

func startComment() string {
	return wrappingComment("START")
}

func finishComment() string {
	return wrappingComment("FINISH")
}

func entriesAsString(entries *hostsFileEntries) *string {
	o := startComment()

	for _, entry := range *entries {
		o = o + fmt.Sprintf("%-16v", *entry.ip) + " " + *entry.hostname
		if entry.ipResovledFrom != nil {
			o = o + "  # IP resolved from " + *entry.ipResovledFrom
		}
		o = o + "\n"
	}

	o = o + finishComment()

	return &o
}
