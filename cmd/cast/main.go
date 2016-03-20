package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/barnybug/go-castv2"
	"github.com/barnybug/go-castv2/controllers"
	"github.com/barnybug/go-castv2/log"
	"github.com/codegangsta/cli"
)

const UrnMedia = "urn:x-cast:com.google.cast.media"

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	commonFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug logging",
		},
		cli.StringFlag{
			Name:  "host",
			Usage: "chromecast hostname or IP (required)",
		},
		cli.IntFlag{
			Name:  "port",
			Usage: "chromecast port",
			Value: 8009,
		},
	}
	app := cli.NewApp()
	app.Name = "cast"
	app.Usage = "Command line tool for the Chromecast"
	app.Version = castv2.Version
	app.Flags = commonFlags
	app.Commands = []cli.Command{
		{
			Name:      "play",
			Usage:     "play some media",
			ArgsUsage: "play url [content type]",
			Action:    cliCommand,
		},
		{
			Name:   "stop",
			Usage:  "stop playing",
			Action: cliCommand,
		},
		{
			Name:   "pause",
			Usage:  "pause playing",
			Action: cliCommand,
		},
		{
			Name:   "volume",
			Usage:  "set current volume",
			Action: cliCommand,
		},
		{
			Name:   "quit",
			Usage:  "close current app on chromecast",
			Action: cliCommand,
		},
		{
			Name:   "script",
			Usage:  "Run the set of commands passed to stdin",
			Action: scriptCommand,
		},
	}
	app.Run(os.Args)
	log.Println("Done")
}

func cliCommand(c *cli.Context) {
	log.Debug = c.GlobalBool("debug")
	if !checkCommand(c.Command.Name, c.Args()) {
		return
	}
	client := connect(c)
	runCommand(client, c.Command.Name, c.Args())
}

func connect(c *cli.Context) *castv2.Client {
	host := c.GlobalString("host")
	log.Printf("Looking up %s...", host)
	ips, err := net.LookupIP(host)
	checkErr(err)

	client := castv2.NewClient()
	err = client.Connect(ips[0], c.GlobalInt("port"))
	checkErr(err)

	log.Println("Connected")
	return client
}

func scriptCommand(c *cli.Context) {
	log.Debug = c.GlobalBool("debug")
	scanner := bufio.NewScanner(os.Stdin)
	commands := [][]string{}

	for scanner.Scan() {
		args := strings.Split(scanner.Text(), " ")
		if len(args) == 0 {
			continue
		}
		if !checkCommand(args[0], args[1:]) {
			return
		}
		commands = append(commands, args)
	}

	client := connect(c)

	for _, args := range commands {
		runCommand(client, args[0], args[1:])
	}
}

var minArgs = map[string]int{
	"play":   1,
	"pause":  0,
	"stop":   0,
	"quit":   0,
	"volume": 1,
}

var maxArgs = map[string]int{
	"play":   2,
	"pause":  0,
	"stop":   0,
	"quit":   0,
	"volume": 1,
}

func checkCommand(cmd string, args []string) bool {
	if _, ok := minArgs[cmd]; !ok {
		fmt.Printf("Command '%s' not understood\n", cmd)
		return false
	}
	if len(args) < minArgs[cmd] {
		fmt.Printf("Command '%s' requires at least %d argument(s)\n", cmd, minArgs[cmd])
		return false
	}
	if len(args) > maxArgs[cmd] {
		fmt.Printf("Command '%s' takes at most %d argument(s)\n", cmd, maxArgs[cmd])
		return false
	}
	return true
}

func runCommand(client *castv2.Client, cmd string, args []string) {
	switch cmd {
	case "play":
		media := client.Media()
		url := args[0]
		contentType := "audio/mpeg"
		if len(args) > 1 {
			contentType = args[1]
		}
		item := controllers.MediaItem{url, "BUFFERED", contentType}
		media.LoadMedia(item, 0, true, map[string]interface{}{}, 5*time.Second)

	case "pause":
		media := client.Media()
		media.Pause(5 * time.Second)

	case "stop":
		if !client.IsPlaying() {
			// if media isn't running, no media can be playing
			return
		}
		media := client.Media()
		media.Stop(5 * time.Second)

	case "volume":
		receiver := client.Receiver()
		level, _ := strconv.ParseFloat(args[0], 64)
		muted := false
		volume := controllers.Volume{Level: &level, Muted: &muted}
		receiver.SetVolume(&volume, 5*time.Second)

	case "quit":
		receiver := client.Receiver()
		receiver.QuitApp(5 * time.Second)

	default:
		fmt.Printf("Command '%s' not understood - ignored\n", cmd)
	}
}
