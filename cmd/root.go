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
			disconnectWithSpinner(p, targetProfile.Server, func() {
				p.Disconnect(targetProfile.ID)
			})
			return
		}

		// disconnect all connections before connecting the target
		if len(conns) > 0 {
			var names []string
			for _, profile := range profiles {
				if _, ok := conns[profile.ID]; ok {
					names = append(names, profile.Server)
				}
			}
			disconnectWithSpinner(p, strings.Join(names, ", "), p.DisconnectAll)
		}

		connectWithSpinner(p, targetProfile)
	},
}

// connectWithSpinner starts the connection and shows a progress line with the
// daemon's current status until the connection is up, fails, or times out.
func connectWithSpinner(p *pritunl.Pritunl, profile pritunl.Profile) {
	sp := newSpinner()
	p.Connect(profile.ID, password())

	deadline := time.Now().Add(30 * time.Second)

	// The daemon registers the connection asynchronously, so an empty
	// status before any status has been observed means pending, not failed.
	seen := false

	for {
		if time.Now().After(deadline) {
			sp.fail(fmt.Sprintf("Connect %s timed out", profile.Server))
			return
		}

		status := p.Connections()[profile.ID].Status
		switch status {
		case "connected":
			sp.ok(fmt.Sprintf("Connected %s", profile.Server))
			return
		case "":
			if seen {
				sp.fail(fmt.Sprintf("Connect %s failed", profile.Server))
				return
			}
		default:
			seen = true
		}

		label := fmt.Sprintf("Connecting %s", profile.Server)
		if status != "" {
			label += color.HiBlackString(" · %s", status)
		}
		sp.tick(label)
		time.Sleep(100 * time.Millisecond)
	}
}

// disconnectWithSpinner runs stop while showing a progress line. The daemon
// answers the stop request only after the tunnel is torn down, so the request
// itself is the wait. Afterwards it polls until no connection is active,
// giving up after 5 seconds.
func disconnectWithSpinner(p *pritunl.Pritunl, names string, stop func()) {
	sp := newSpinner()
	label := "Disconnecting " + names

	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()

	deadline := time.Now().Add(30 * time.Second)
	for {
		select {
		case <-done:
			deadline = time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				if !anyActive(p) {
					sp.ok("Disconnected " + names)
					return
				}
				sp.tick(label)
				time.Sleep(100 * time.Millisecond)
			}
			sp.fail("Disconnect " + names + " timed out")
			return
		default:
			if time.Now().After(deadline) {
				sp.fail("Disconnect " + names + " timed out")
				return
			}
			sp.tick(label)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// anyActive reports whether the daemon still has a connection that is not
// disconnected.
func anyActive(p *pritunl.Pritunl) bool {
	for _, conn := range p.Connections() {
		if conn.Status != "disconnected" {
			return true
		}
	}
	return false
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
