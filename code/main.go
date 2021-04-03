package main

/*
 * Reacts to sh.keptn.event.evaluation.finished and sh.keptn.event.remediation.finished
 */

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	"github.com/kelseyhightower/envconfig"

	keptn "github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

var keptnOptions = keptn.KeptnOpts{}

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int `envconfig:"RCV_PORT" default:"8080"`
	// Path to which cloudevents are sent
	Path string `envconfig:"RCV_PATH" default:"/"`
	// Whether we are running locally (e.g., for testing) or on production
	Env string `envconfig:"ENV" default:"local"`
	// URL of the Keptn configuration service (this is where we can fetch files from the config repo)
	ConfigurationServiceUrl string `envconfig:"CONFIGURATION_SERVICE" default:""`
}

type ZendeskDetails struct {
	BaseURL              string
	EndUserEmail         string
	APIToken             string
	TicketForProblems    bool
	TicketForEvaluations bool
}

type KeptnDetails struct {
	Domain    string
	BridgeURL string
}

var ZENDESK_DETAILS ZendeskDetails
var KEPTN_DETAILS KeptnDetails

// ServiceName specifies the current services name (e.g., used as source when sending CloudEvents)
const ServiceName = "zendesk-service"

// This method gets called when a new event is received from the Keptn Event Distributor
func processKeptnCloudEvent(ctx context.Context, event cloudevents.Event) error {

	// create keptn handler
	log.Printf("[main.go] Initializing Keptn Handler")
	myKeptn, err := keptnv2.NewKeptn(&event, keptnOptions)
	if err != nil {
		return errors.New("Could not create Keptn Handler: " + err.Error())
	}

	setupAndDebug(myKeptn, event)

	switch event.Type() {

	// Listen for remediation.finished
	case keptnv2.GetFinishedEventType(keptnv2.RemediationTaskName): // sh.keptn.event.remediation.finished
		log.Printf("Processing Remediation.Finished Event")

		eventData := &keptnv2.RemediationFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		HandleRemediationFinishedEvent(myKeptn, event, eventData)

	// Handle evaluation.finished event type
	case keptnv2.GetFinishedEventType(keptnv2.EvaluationTaskName): // sk.keptn.event.evaluation.finished
		log.Printf("Processing Evaluation.Finished Event")

		eventData := &keptnv2.EvaluationFinishedEventData{}
		parseKeptnCloudEventPayload(event, eventData)

		HandleEvaluationFinishedEvent(myKeptn, event, eventData)
	}

	return nil

}

/**
 * Usage: ./main
 * no args: starts listening for cloudnative events on localhost:port/path
 *
 * Environment Variables
 * env=runlocal   -> will fetch resources from local drive instead of configuration service
 */
func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("[main.go] Failed to process env var: %s", err)
	}

	os.Exit(_main(os.Args[1:], env))
}

/**
 * Opens up a listener on localhost:port/path and passes incoming requets to gotEvent
 */
func _main(args []string, env envConfig) int {
	// configure keptn options
	if env.Env == "local" {
		log.Println("[main.go] env=local: Running with local filesystem to fetch resources")
		keptnOptions.UseLocalFileSystem = true
	}

	keptnOptions.ConfigurationServiceURL = env.ConfigurationServiceUrl

	log.Printf("[main.go] Starting %s...", ServiceName)
	log.Printf("[main.go]     on Port = %d; Path=%s", env.Port, env.Path)

	ctx := context.Background()
	ctx = cloudevents.WithEncodingStructured(ctx)

	log.Printf("[main.go] Creating new http handler")

	// configure http server to receive cloudevents
	p, err := cloudevents.NewHTTP(cloudevents.WithPath(env.Path), cloudevents.WithPort(env.Port))

	if err != nil {
		log.Fatalf("[main.go] failed to create client, %v", err)
	}
	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("[main.go] Starting receiver")
	log.Fatal(c.StartReceiver(ctx, processKeptnCloudEvent))

	return 0
}

/**
 * Parses a Keptn Cloud Event payload (data attribute)
 */
func parseKeptnCloudEventPayload(event cloudevents.Event, data interface{}) error {
	err := event.DataAs(data)
	if err != nil {
		log.Fatalf("Got Data Error: %s", err.Error())
		return err
	}
	return nil
}

func setZendeskDetails() {
	ZENDESK_DETAILS.BaseURL = os.Getenv("ZENDESK_BASE_URL")
	ZENDESK_DETAILS.EndUserEmail = os.Getenv("ZENDESK_END_USER_EMAIL")
	ZENDESK_DETAILS.APIToken = os.Getenv("ZENDESK_API_TOKEN")
	ZENDESK_DETAILS.TicketForProblems, _ = strconv.ParseBool(os.Getenv("ZENDESK_TICKET_FOR_PROBLEMS"))
	ZENDESK_DETAILS.TicketForEvaluations, _ = strconv.ParseBool(os.Getenv("ZENDESK_TICKET_FOR_EVALUATIONS"))
}

func setKeptnDetails() {
	KEPTN_DETAILS.Domain = os.Getenv("KEPTN_DOMAIN")

	// If Bridge URL isn't set in YAML file, default to the KEPTN_DOMAIN which is mandatory
	if os.Getenv("KEPTN_BRIDGE_URL") == "" {
		KEPTN_DETAILS.BridgeURL = os.Getenv("KEPTN_DOMAIN")
	} else {
		KEPTN_DETAILS.BridgeURL = os.Getenv("KEPTN_BRIDGE_URL")
	}
}

func setupAndDebug(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event) {
	log.Printf("[main.go] gotEvent(%s): %s - %s", incomingEvent.Type(), myKeptn.KeptnContext, incomingEvent.Context.GetID())

	// Get Debug Mode
	// This is set in the service.yaml as DEBUG "true"
	DEBUG, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	log.Printf("[main.go] Debug Mode: %v \n", DEBUG)

	// Set Zendesk Details
	setZendeskDetails()

	// Get Dynatrace Tenant
	dynaTraceTenant := os.Getenv("DT_TENANT")

	// KEPTN_DOMAIN must be set but KEPTN_BRIDGE_URL is optional in zendesk-service deployment.yaml file
	setKeptnDetails()

	if ZENDESK_DETAILS.BaseURL == "" ||
		KEPTN_DETAILS.Domain == "" {
		log.Println("[main.go] Missing mandatory input parameters ZENDESK_DETAILS and / or KEPTN_DOMAIN.")
	}

	if DEBUG {
		log.Println("[main.go] --- Printing Zendesk Input Details ---")
		log.Printf("[main.go] Base URL: %s \n", ZENDESK_DETAILS.BaseURL)
		log.Printf("[main.go] Ticket For Problems: %v \n", ZENDESK_DETAILS.TicketForProblems)
		log.Printf("[main.go] Ticket For Problems: %v \n", ZENDESK_DETAILS.TicketForEvaluations)
		log.Println("[main.go] --- End Printing Zendesk Input Details ---")

		log.Printf("[main.go] Dynatrace Tenant: %s \n", dynaTraceTenant)
		log.Printf("[main.go] Keptn Domain: %s \n", KEPTN_DETAILS.Domain)
		log.Printf("[main.go] Keptn Bridge URL: %s \n", KEPTN_DETAILS.BridgeURL)

		// At this point, we have all mandatory input params. Proceed
		log.Println("[main.go] Got all input variables. Proceeding...")

		if ZENDESK_DETAILS.TicketForProblems {
			log.Println("[main.go] Will create tickets for problems")
		} else {
			log.Println("[main.go] Will NOT create tickets for problems")
		}

		if ZENDESK_DETAILS.TicketForEvaluations {
			log.Println("[main.go] Will create tickets for evaluations")
		} else {
			log.Println("[main.go] Will NOT create tickets for evaluations")
		}

		SEND_EVENT, _ := strconv.ParseBool(os.Getenv("SEND_EVENT"))
		if SEND_EVENT {
			log.Println("[main.go] Will send events")
		} else {
			log.Println("[main.go] Will NOT send events")
		}
	}
}
