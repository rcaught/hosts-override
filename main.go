package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type hostsFileEntry struct {
	hostname       *string
	ip             *string
	ipResovledFrom *string
}

type hostsFileEntries []*hostsFileEntry

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
			hostsFileLocation := hostsFileLocation()
			createHostsBackup(hostsFileLocation)
			removeOverrides(hostsFileLocation) // Fixes unclean shutdown
			parseAndAppend(entries, hostsFileLocation, &refresh, &refreshInterval)

			if refresh {
				refreshTicker := time.NewTicker(refreshInterval)

				go func() {
					for {
						select {
						case <-refreshTicker.C:
							removeOverrides(hostsFileLocation)
							clearScreen()
							parseAndAppend(entries, hostsFileLocation, &refresh, &refreshInterval)
						}
					}
				}()
			}

			waitUntilExit()
			removeOverrides(hostsFileLocation)
		},
	}

	rootCmd.Flags().BoolVarP(&refresh, "refresh", "r", false, "Refresh unresolved hosts")
	rootCmd.Flags().DurationVarP(&refreshInterval, "refresh-interval", "i", time.Duration(300)*time.Second, "Refresh Interval")

	return rootCmd
}

func main() {
	overrideCmd().Execute()
}

func clearScreen() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func parseAndAppend(entries *hostsFileEntries, hostsFileLocation *string, refresh *bool, refreshInterval *time.Duration) {
	fmt.Println("\nhosts-override: Overriding hosts file entries for the lifetime of the process")

	parsedOverrides := parsedOverrides(entries, refresh)

	if parsedOverrides == nil {
		return
	}

	parsedOverridesForHosts := parsedOverridesForHosts(parsedOverrides)
	appendOverrides(hostsFileLocation, parsedOverridesForHosts)
	if *refresh == true {
		fmt.Println("\n(Refreshing every " + refreshInterval.String() + ")...")
	}
	fmt.Println("\n" + *parsedOverridesAsHosts(parsedOverrides) + "\n")
	fmt.Println("\nPress CTRL-C to exit gracefully (hosts file will reset)")
}

func parseArgs(args *[]string) *hostsFileEntries {
	var entries hostsFileEntries

	for _, pair := range *args {
		hv := strings.Split(pair, ",")
		entries = append(entries, &hostsFileEntry{&hv[0], &hv[1], nil})
	}

	return &entries
}

func hostsFileLocation() *string {
	var hostsFile string

	if runtime.GOOS == "windows" {
		hostsFile = "${SystemRoot}/System32/drivers/etc/hosts"
	} else {
		hostsFile = "/etc/hosts"
	}

	return &hostsFile
}

func appendOverrides(hostsFileLocation *string, parsedOverridesAsHosts *string) {
	f, err := os.OpenFile(*hostsFileLocation, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err)
	}

	defer f.Close()

	if _, err := f.WriteString(*parsedOverridesAsHosts); err != nil {
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

func wrappingComment(custom string) string {
	return "\n\n#########################\n" +
		"# hosts-override " + custom +
		"\n#########################\n\n"
}

func startComment() string {
	return wrappingComment("START")
}

func finishComment() string {
	return wrappingComment("FINISH")
}

func parsedOverridesForHosts(entries *hostsFileEntries) *string {
	o := startComment()
	o = o + *parsedOverridesAsHosts(entries)
	o = o + finishComment()

	return &o
}

func parsedOverridesAsHosts(entries *hostsFileEntries) *string {
	var o string
	for _, entry := range *entries {
		o = o + *entry.ip + " " + *entry.hostname + " # IP resolved from " + *entry.ipResovledFrom + "\n"
	}

	return &o
}

func createHostsBackup(hostsFileLocation *string) {
	input, err := ioutil.ReadFile(*hostsFileLocation)
	if err != nil {
		fmt.Println(err)
		return
	}

	backupLocation := *hostsFileLocation + ".backup-" + strconv.FormatInt(time.Now().Unix(), 10)

	err = ioutil.WriteFile(backupLocation, input, 0644)
	if err != nil {
		fmt.Println("Error creating", backupLocation)
		fmt.Println(err)
		return
	}
}

func parsedOverrides(entries *hostsFileEntries, refresh *bool) *hostsFileEntries {
	expandedEntries := hostsFileEntries{}

	for _, entry := range *entries {
		if maybeIP := *maybeIP(entry.ip); maybeIP != "" {
			expandedEntries = append(expandedEntries, entry)
		} else {
			ips, err := net.LookupIP(*entry.ip)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
				if *refresh {
					return nil
				}
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

func maybeIP(value *string) *string {
	v := net.ParseIP(*value)

	four := v.To4()
	six := v.To16()
	maybeIP := ""

	if four != nil {
		maybeIP = four.String()
	} else if six != nil {
		maybeIP = six.String()
	}

	return &maybeIP
}

func waitUntilExit() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)

		done <- true
	}()

	<-done
}
