package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func overrideCmd() *cobra.Command {
	var hosts []string
	var values []string

	rootCmd := &cobra.Command{
		Use:   "hosts-override",
		Short: "Override hosts file entries for the life of the process",
		Run: func(cmd *cobra.Command, args []string) {
			hostsFileLocation := hostsFileLocation()
			createHostsBackup(hostsFileLocation)
			parsedOverrides := parsedOverrides(&hosts, &values)
			parsedOverridesAsHosts := parsedOverridesAsHosts(parsedOverrides)
			appendOverrides(hostsFileLocation, parsedOverridesAsHosts)
			waitUntilExit()
			removeOverrides(hostsFileLocation, parsedOverridesAsHosts)
		},
	}

	rootCmd.Flags().StringSliceVarP(&hosts, "host", "0", []string{}, "Host name to override (can be multiple)")
	rootCmd.Flags().StringSliceVarP(&values, "value", "1", []string{}, "IP or unresolved host name (can be multiple)")

	return rootCmd
}

func main() {
	overrideCmd().Execute()
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

func removeOverrides(hostsFileLocation *string, parsedOverridesAsHosts *string) {
	contents, err := ioutil.ReadFile(*hostsFileLocation)

	if err != nil {
		fmt.Println(err)
		return
	}

	removedOverrides := strings.Replace(string(contents), *parsedOverridesAsHosts, "", 1)

	if err := ioutil.WriteFile(*hostsFileLocation, []byte(removedOverrides), 0); err != nil {
		log.Println(err)
	}
}

func parsedOverridesAsHosts(parsedOverrides *map[string][]string) *string {
	o := "\n\n#########################\n" +
		"# hosts-override START  #\n" +
		"#########################\n\n"

	for ip, hosts := range *parsedOverrides {
		o = o + ip + " "
		for _, host := range hosts {
			o = o + host + " "
		}

		o = o + "\n"
	}

	o = o + "\n#########################\n" +
		"# hosts-override FINISH #\n" +
		"#########################\n"

	return &o
}

func createHostsBackup(hostsFileLocation *string) {
	input, err := ioutil.ReadFile(*hostsFileLocation)
	if err != nil {
		fmt.Println(err)
		return
	}

	backupLocation := *hostsFileLocation + ".backup"

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

	fmt.Println("\nPress CTRL-C to exit")
	<-done
}
