package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/shawnpeng17/17vpn/internal/pritunl"
)

var rootCmd = &cobra.Command{
	Use:   "17vpn",
	Short: "17vpn tool",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			color.Red(err.Error())
			return
		}

		p := pritunl.New()
		profiles := p.Profiles()
		conns := p.Connections()

		var id string
		if len(args) == 1 {
			id = args[0]
		} else {
			if err := list(profiles, conns); err != nil {
				color.Yellow(err.Error())
				return
			}

			fmt.Println()

			var options []string
			for _, profile := range profiles {
				options = append(options, profile.Server)
			}
			prompt := &survey.Input{
				Message:       "Enter ID or Server",
				Default:       "",
			}
			if err := survey.AskOne(prompt, &id); err != nil {
				color.Red(err.Error())
				return
			}
		}

		if id == "" {
			return
		}

		// check profile exist
		var targetProfile pritunl.Profile
		isActionDisconnect := false
		for i, profile := range profiles {
			if strconv.Itoa(i+1) == id || strings.ToUpper(id) == profile.Server {
				targetProfile = profile
				isActionDisconnect = conns[profile.ID].Status == "connected"
				break
			}
		}
		if targetProfile == (pritunl.Profile{}) {
			color.Red("Profile not exists!")
			return
		}

		if isActionDisconnect {
			color.White("Disconnecting %s...", targetProfile.Server)
			p.Disconnect(targetProfile.ID)
			return
		}

		// disconnect all connections before connecting the target
		if len(conns) > 0 {
			for _, profile := range profiles {
				if _, ok := conns[profile.ID]; ok {
					color.White("Disconnecting %s...", profile.Server)
				}
			}
			p.DisconnectAll()
			waitDisconnected(p, 5*time.Second)
		}

		// connect target profile
		color.Yellow("Connecting %s...", targetProfile.Server)
		p.Connect(targetProfile.ID, password())

		timeout := time.NewTimer(30 * time.Second)

		// The daemon registers the connection asynchronously, so an empty
		// status before any status has been observed means pending, not failed.
		seen := false

	Loop:
		for {
			select {
			case <-timeout.C:
				color.Red("Connect %s timeout!", targetProfile.Server)
				break Loop
			default:
				status := p.Connections()[targetProfile.ID].Status
				switch status {
				case "connected":
					color.Green("Connect %s completed!", targetProfile.Server)
					break Loop
				case "":
					if seen {
						color.Red("Connect %s failed!", targetProfile.Server)
						break Loop
					}
				default:
					seen = true
				}
				time.Sleep(200 * time.Millisecond)
			}
		}
	},
}

// waitDisconnected polls the daemon until every connection is gone or has
// status "disconnected", or until timeout passes.
func waitDisconnected(p *pritunl.Pritunl, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		active := false
		for _, conn := range p.Connections() {
			if conn.Status != "disconnected" {
				active = true
				break
			}
		}
		if !active {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		return
	}
}
