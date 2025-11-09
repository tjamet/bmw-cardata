package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	bmwcardata "github.com/tjamet/bmw-cardata"
)

func dumpOutput(data any, err error) error {
	if err != nil {
		return err
	}
	e := json.NewEncoder(os.Stdout)
	return e.Encode(data)
}

func main() {
	defaultSessionPath, err := bmwcardata.DefaultSessionPath()
	if err != nil {
		log.Fatal(err)
	}
	defaultFrom := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	defaultTo := time.Now().Format("2006-01-02")

	ctx := context.Background()

	sessionPath := flag.String("session-path", defaultSessionPath, "Path to the session file")
	clientID := flag.String("client-id", "", "Client ID")
	vin := flag.String("vin", "", "VIN")

	from := flag.String("from", defaultFrom, "From date (YYYY-MM-DD)")
	to := flag.String("to", defaultTo, "To date (YYYY-MM-DD)")

	nextToken := flag.String("next-token", "", "Next token")

	containerID := flag.String("container-id", "", "Container ID")

	archivePath := flag.String("archive-path", "", "Archive path")

	newClient := func() *bmwcardata.Client {
		client, err := bmwcardata.NewClient(
			bmwcardata.WithAuthenticator(bmwcardata.Must(bmwcardata.NewAuthenticator(
				bmwcardata.WithSessionStore(&bmwcardata.FileSessionStore{Path: *sessionPath}),
				bmwcardata.WithClientID(*clientID),
				bmwcardata.WithPromptURI(func(uri, code, complete string) {
					fmt.Println("Open the following URL in your browser:")
					fmt.Println(complete)
				}),
			))),
		)
		if err != nil {
			log.Fatal(err)
		}
		return client
	}

	commands := map[string]func(ctx context.Context) error{
		"mappings": func(ctx context.Context) error {
			return dumpOutput(newClient().GetMappings(ctx))
		},
		"get-basic-data": func(ctx context.Context) error {
			return dumpOutput(newClient().GetBasicData(ctx, *vin))
		},
		"get-charging-history": func(ctx context.Context) error {
			from, err := time.Parse("2006-01-02", *from)
			if err != nil {
				return err
			}
			to, err := time.Parse("2006-01-02", *to)
			if err != nil {
				return err
			}
			options := []bmwcardata.GetChargingHistoryParamsOption{}
			if *nextToken != "" {
				options = append(options, bmwcardata.WithChargingHistoryNextToken(*nextToken))
			}
			return dumpOutput(newClient().GetChargingHistory(ctx, *vin, from, to, options...))
		},
		"get-image": func(ctx context.Context) error {
			return dumpOutput(newClient().GetImage(ctx, *vin))
		},
		"get-location-based-charging-settings": func(ctx context.Context) error {
			options := []bmwcardata.GetLocationBasedChargingSettingsParamsOption{}
			if *nextToken != "" {
				options = append(options, bmwcardata.WithLocationBasedChargingSettingsNextToken(*nextToken))
			}
			return dumpOutput(newClient().GetLocationBasedChargingSettings(ctx, *vin, options...))
		},
		"get-smart-maintenance-tyre-diagnosis": func(ctx context.Context) error {
			return dumpOutput(newClient().GetSmartMaintenanceTyreDiagnosis(ctx, *vin))
		},
		"list-containers": func(ctx context.Context) error {
			return dumpOutput(newClient().ListContainers(ctx))
		},
		"get-container-details": func(ctx context.Context) error {
			return dumpOutput(newClient().GetContainerDetails(ctx, *containerID))
		},
		"delete-container": func(ctx context.Context) error {
			return dumpOutput(newClient().DeleteContainer(ctx, *containerID))
		},
		"get-telematic-data": func(ctx context.Context) error {
			return dumpOutput(newClient().GetTelematicData(ctx, *vin, *containerID))
		},
		"read-archive": func(ctx context.Context) error {
			return dumpOutput(bmwcardata.ReadArchive(*archivePath))
		},
		"stream-telematic-data": func(ctx context.Context) error {
			client := newClient()
			err := client.StartEventStream()
			if err != nil {
				return err
			}
			defer client.StopEventStream()
			e := json.NewEncoder(os.Stdout)
			client.Subscribe(ctx, *vin, func(message bmwcardata.StreamedMessage) {
				err := e.Encode(message)
				if err != nil {
					log.Fatal(err)
				}
			})
			<-client.Done()
			return nil
		},
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [flags] <command>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		for command := range commands {
			fmt.Fprintf(os.Stderr, "  %s\n", command)
		}
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	for _, method := range flag.Args() {
		command, ok := commands[method]
		if !ok {
			log.Fatalf("Unknown command: %s", method)
		}
		err := command(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
