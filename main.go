package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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
			hosts, values := parseArgs(&args)
			hostsFileLocation := hostsFileLocation()
			createHostsBackup(hostsFileLocation)
			removeOverrides(hostsFileLocation) // Fixes unclean shutdown

			fmt.Println("\nOverriding hosts file entries for the lifetime of the process:")
			parseAndAppend(hosts, values, hostsFileLocation, &refresh, &refreshInterval)

			if refresh == true {
				refreshTicker := time.NewTicker(refreshInterval)

				go func() {
					for {
						select {
						case <-refreshTicker.C:
							removeOverrides(hostsFileLocation)
							parseAndAppend(hosts, values, hostsFileLocation, &refresh, &refreshInterval)
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

func parseAndAppend(hosts *[]string, values *[]string, hostsFileLocation *string, refresh *bool, refreshInterval *time.Duration) {
	parsedOverrides := parsedOverrides(hosts, values)
	parsedOverridesForHosts := parsedOverridesForHosts(parsedOverrides)
	appendOverrides(hostsFileLocation, parsedOverridesForHosts)
	if *refresh == true {
		fmt.Println("\n(Refreshing in " + refreshInterval.String() + ")...")
	}
	fmt.Println("\n" + *parsedOverridesAsHosts(parsedOverrides) + "\n")
	fmt.Println("\nPress CTRL-C to exit")
}

func parseArgs(args *[]string) (*[]string, *[]string) {
	var hosts []string
	var values []string

	for _, pair := range *args {
		hv := strings.Split(pair, ",")
		hosts = append(hosts, hv[0])
		values = append(values, hv[1])
	}

	return &hosts, &values
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

func parsedOverridesForHosts(parsedOverrides *map[string][]string) *string {
	o := startComment()
	o = o + *parsedOverridesAsHosts(parsedOverrides)
	o = o + finishComment()

	return &o
}

func parsedOverridesAsHosts(parsedOverrides *map[string][]string) *string {
	var o string
	for ip, hosts := range *parsedOverrides {
		o = o + ip + " "
		for _, host := range hosts {
			o = o + host + " "
		}

		o = o + "\n"
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

func parsedOverrides(hosts *[]string, values *[]string) *map[string][]string {
	parsed := map[string][]string{}

	for i, host := range *hosts {
		value := (*values)[i]

		if maybeIP := *maybeIP(&value); maybeIP != "" {
			parsed[maybeIP] = append(parsed[maybeIP], host)
		} else {
			ips, err := net.LookupIP(value)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
				os.Exit(1)
			}

			for _, ip := range ips {
				ip := ip.String()

				parsed[ip] = append(parsed[ip], host)
			}

		}
	}

	return &parsed
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
