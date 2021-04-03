package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

func HandleEvaluationFinishedEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data *keptnv2.EvaluationFinishedEventData) {
	log.Println("[eventhandlers.go] Handling evaluation.finished Event:", incomingEvent.Context.GetID())

	if !ZENDESK_DETAILS.TicketForEvaluations {
		log.Println("[eventhandlers.go] TicketForEvaluations flag is set to false. Got an evaluation.finished from Keptn but doing nothing. If you want a ticket, set flag to true")
		return
	}

	ticketKey := createZendeskTicketForEvaluationFinished(myKeptn, data)
	ticketURL := ZENDESK_DETAILS.BaseURL + "/agent/tickets/" + ticketKey

	// If the SEND_EVENT flag is set in service.yaml send an event to the relevant tool
	SEND_EVENT, _ := strconv.ParseBool(os.Getenv("SEND_EVENT"))
	if SEND_EVENT {
		sendEventForEvaluationFinishedEvents("dynatrace", "CUSTOM_INFO", ticketURL, data, myKeptn)
	}
}

func HandleRemediationFinishedEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data *keptnv2.RemediationFinishedEventData) {
	log.Printf("[eventhandlers.go] Handling remediation.finished event: %s", incomingEvent.Context.GetID())

	if !ZENDESK_DETAILS.TicketForProblems {
		log.Println("[eventhandlers.go] TicketForProblems flag is set to false. Got a remediation.finished from Keptn but doing nothing. If you want a ticket, set flag to true")
		return
	}

	ticketKey := createZendeskTicketForRemediationFinished(myKeptn, data)
	ticketURL := ZENDESK_DETAILS.BaseURL + "/agent/tickets/" + ticketKey

	// If the SEND_EVENT flag is set in service.yaml send an event to the relevant tool
	SEND_EVENT, _ := strconv.ParseBool(os.Getenv("SEND_EVENT"))
	if SEND_EVENT {
		sendEventForRemediationFinishedEvents("dynatrace", "CUSTOM_INFO", ticketURL, data, myKeptn)
	}
}

//*******************************
//       Helper functions
//*******************************

/********************************************
*   REMEDIATION.FINISHED SPECIFIC METHODS
*********************************************/

func createCustomPropertiesForRemediationFinishedEvents(myKeptn *keptnv2.Keptn, data *keptnv2.RemediationFinishedEventData, ticketURL string) map[string]string {
	var customProperties = make(map[string]string)

	customProperties["Result"] = string(data.Result)
	customProperties["Keptn Project"] = data.EventData.GetProject()
	customProperties["Keptn Service"] = data.EventData.GetService()
	customProperties["Keptn Stage"] = data.EventData.GetStage()
	customProperties["Ticket"] = ticketURL
	customProperties["SentBy"] = "Keptn"
	bridgeURL := KEPTN_DETAILS.BridgeURL + "/project/" + data.EventData.GetProject() + "/sequence/" + myKeptn.KeptnContext
	customProperties["BridgeURL"] = bridgeURL

	return customProperties
}

// This function relies on standard keptn tags:
// keptn_project, keptn_service and keptn_stage being present
//
// Note: This method might be replaced in future if we can send events that the dynatrace-service consumes
// As the dynatrace-service contains nice helper methods to send events.
func sendEventForRemediationFinishedEvents(eventDestination string, eventType string, ticketURL string, data *keptnv2.RemediationFinishedEventData, myKeptn *keptnv2.Keptn) {
	log.Println("[eventhandlers.go] Sending event to:", eventDestination, " as type:", eventType)

	// Split ticketURL by last forward slash to get the ticket Key
	ticketKey := ticketURL[strings.LastIndex(ticketURL, "/")+1:]

	// Send Dynatrace Event
	if eventDestination == "dynatrace" && os.Getenv("DT_TENANT") != "" && os.Getenv("DT_API_TOKEN") != "" {

		dynatraceTenant := os.Getenv("DT_TENANT")
		dynatraceAPIToken := os.Getenv("DT_API_TOKEN")
		dynatraceAPITokenHeader := "Api-Token " + dynatraceAPIToken

		// Build data
		var dtInfoEvent = new(DtInfoEvent)
		dtInfoEvent.EventType = eventType
		dtInfoEvent.Source = ServiceName
		dtInfoEvent.Title = "Ticket Created: #" + ticketKey
		dtInfoEvent.AttachRules = createAttachRulesForRemediationFinishedEvents(data)
		dtInfoEvent.Description = "Keptn Remediation Attempt"
		customProperties := createCustomPropertiesForRemediationFinishedEvents(myKeptn, data, ticketURL)
		dtInfoEvent.CustomProperties = customProperties

		//Encode the data
		jsonString, _ := json.Marshal(dtInfoEvent)

		client := &http.Client{}

		dtTenantURL := "https://" + dynatraceTenant + "/api/v1/events"
		req, _ := http.NewRequest("POST", dtTenantURL, bytes.NewReader(jsonString))
		req.Header.Add("accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", dynatraceAPITokenHeader)

		// Send Request
		resp, err := client.Do(req)

		//Handle Error
		if err != nil {
			log.Fatalf("[eventhandlers.go] An Error Occured Sending POSt to Zendesk %v", err)
		}
		defer resp.Body.Close()

		//Read the response body
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

	}

}

func createZendeskTicketForRemediationFinished(myKeptn *keptnv2.Keptn, data *keptnv2.RemediationFinishedEventData) string {

	log.Println("[eventhandlers.go] Creating Zendesk Body details for remediation.finished...")

	// Build title field
	title := "[REMEDIATION] " + data.GetProject() + " - " + data.GetService() + " - " + data.GetStage() + " - Result: " + string(data.Result)

	// TODO - Zendesk doesn't accept Markdown by default. Reformat this...
	bodyContent := "||*Remediation Status*||*Project*||*Service*||*Stage*||\n"
	stringResult := string(data.Result)
	result := ""
	if stringResult == "pass" {
		result += stringResult + " ✅"
	} else if stringResult == "warning" {
		result += stringResult + " ⚠"
	} else if stringResult == "fail" {
		result += stringResult + " ❌"
	} else {
		result += stringResult
	}
	bodyContent += "|" + result + "|" + data.GetProject() + "|" + data.GetService() + "|" + data.GetStage() + "|\n\n"

	// Add Message
	bodyContent += "Message: " + data.Message + "\n\n"

	// Add Keptn Context
	bodyContent += "Keptn Context ID: " + myKeptn.KeptnContext + "\n"

	// Add link to Keptn Bridge
	bridgeURL := KEPTN_DETAILS.BridgeURL + "/project/" + data.EventData.GetProject() + "/sequence/" + myKeptn.KeptnContext
	bodyContent += "[Link To Keptn's Bridge|" + bridgeURL + "]"

	// Build map of labels which we take from the cloudevent, which we then attach to the Zendesk ticket
	labels := createZendeskLabelsForRemediationFinishedEvents(data)

	// Send the POST to Zendesk
	ticketKey := createZendeskTicket(title, bodyContent, labels)
	return ticketKey
}

func createZendeskLabelsForRemediationFinishedEvents(data *keptnv2.RemediationFinishedEventData) []string {
	//[]string{"foo:bar", "this:that"}
	labels := []string{}

	// Add Keptn Project, Service and Stage as labels
	// Zendesk labels don't accept spaces so convert spaces to dashes
	value := strings.ReplaceAll(data.EventData.GetProject(), " ", "-")
	labels = append(labels, "keptn_project:"+value)

	value = strings.ReplaceAll(data.EventData.GetService(), " ", "-")
	labels = append(labels, "keptn_service:"+value)

	value = strings.ReplaceAll(data.EventData.GetStage(), " ", "-")
	labels = append(labels, "keptn_service:"+value)

	// Add result as a label (pass, warning or fail)
	labels = append(labels, "keptn_result:"+string(data.Result))

	for labelKey, labelValue := range data.Labels {
		// Replace spaces with dashes for the Key and Value
		labelKeyClean := strings.ReplaceAll(labelKey, " ", "-")
		labelValueClean := strings.ReplaceAll(labelValue, " ", "-")

		//Stick the cleaned key and value back together
		cleanKeyValueLabel := fmt.Sprint(labelKeyClean, ":", labelValueClean)

		labels = append(labels, cleanKeyValueLabel) // Append this "key":"value" using Sprint so as to not add spaces
	}

	return labels
}

func createAttachRulesForRemediationFinishedEvents(data *keptnv2.RemediationFinishedEventData) DtAttachRules {
	attachRule := DtAttachRules{
		TagRule: []DtTagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []DtTag{
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   data.GetProject(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   data.GetStage(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   data.GetService(),
					},
				},
			},
		},
	}

	return attachRule
}

/********************************************
*   EVALUATION.FINISHED SPECIFIC METHODS
*********************************************/

// This function relies on standard keptn tags:
// keptn_project, keptn_service and keptn_stage being present
//
// Note: This method might be replaced in future if we can send events that the dynatrace-service consumes
// As the dynatrace-service contains nice helper methods to send events.
func sendEventForEvaluationFinishedEvents(eventDestination string, eventType string, ticketURL string, data *keptnv2.EvaluationFinishedEventData, myKeptn *keptnv2.Keptn) {
	log.Println("[eventhandlers.go] Sending event to:", eventDestination, " as type:", eventType)

	// Split ticketURL by last forward slash to get the ticket Key
	ticketKey := ticketURL[strings.LastIndex(ticketURL, "/")+1:]

	// Send Dynatrace Event
	if eventDestination == "dynatrace" && os.Getenv("DT_TENANT") != "" && os.Getenv("DT_API_TOKEN") != "" {

		dynatraceTenant := os.Getenv("DT_TENANT")
		dynatraceAPIToken := os.Getenv("DT_API_TOKEN")
		dynatraceAPITokenHeader := "Api-Token " + dynatraceAPIToken

		// Build data
		var dtInfoEvent = new(DtInfoEvent)
		dtInfoEvent.EventType = eventType
		dtInfoEvent.Source = ServiceName
		dtInfoEvent.Title = "Ticket Created: #" + ticketKey
		dtInfoEvent.AttachRules = createAttachRulesForEvaluationFinishedEvents(data)
		dtInfoEvent.Description = "Keptn Quality Gate Evaluation"
		customProperties := createCustomPropertiesForEvaluationFinishedEvents(myKeptn, data, ticketURL)
		dtInfoEvent.CustomProperties = customProperties

		//Encode the data
		jsonString, _ := json.Marshal(dtInfoEvent)

		client := &http.Client{}

		dtTenantURL := "https://" + dynatraceTenant + "/api/v1/events"
		req, _ := http.NewRequest("POST", dtTenantURL, bytes.NewReader(jsonString))
		req.Header.Add("accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", dynatraceAPITokenHeader)

		// Send Request
		resp, err := client.Do(req)

		//Handle Error
		if err != nil {
			log.Fatalf("[eventhandlers.go] An Error Occured Sending POST to Zendesk %v", err)
		}
		defer resp.Body.Close()

		//Read the response body
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

	}

}

func createAttachRulesForEvaluationFinishedEvents(data *keptnv2.EvaluationFinishedEventData) DtAttachRules {
	attachRule := DtAttachRules{
		TagRule: []DtTagRule{
			{
				MeTypes: []string{"SERVICE"},
				Tags: []DtTag{
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_project",
						Value:   data.EventData.GetProject(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_stage",
						Value:   data.EventData.GetStage(),
					},
					{
						Context: "CONTEXTLESS",
						Key:     "keptn_service",
						Value:   data.EventData.GetService(),
					},
				},
			},
		},
	}

	return attachRule
}

func createCustomPropertiesForEvaluationFinishedEvents(myKeptn *keptnv2.Keptn, data *keptnv2.EvaluationFinishedEventData, ticketURL string) map[string]string {
	var customProperties = make(map[string]string)
	//customProperties = make(map[string]string)

	customProperties["Quality Gate Result"] = data.Evaluation.Result
	customProperties["Quality Gate Score"] = fmt.Sprint(data.Evaluation.Score)
	customProperties["Keptn Project"] = data.EventData.GetProject()
	customProperties["Keptn Service"] = data.EventData.GetService()
	customProperties["Keptn Stage"] = data.EventData.GetStage()
	customProperties["Ticket"] = ticketURL
	customProperties["SentBy"] = "Keptn"
	bridgeURL := KEPTN_DETAILS.BridgeURL + "/project/" + data.EventData.GetProject() + "/sequence/" + myKeptn.KeptnContext
	customProperties["BridgeURL"] = bridgeURL

	return customProperties
}

func createZendeskLabelsForEvaluationFinishedEvents(data *keptnv2.EvaluationFinishedEventData) []string {
	//[]string{"foo:bar", "this:that"}
	labels := []string{}

	// Add Keptn Project, Service and Stage as labels
	// Zendesk labels don't accept spaces so convert spaces to dashes
	value := strings.ReplaceAll(data.EventData.GetProject(), " ", "-")
	labels = append(labels, "keptn_project:"+value)

	value = strings.ReplaceAll(data.EventData.GetService(), " ", "-")
	labels = append(labels, "keptn_service:"+value)

	value = strings.ReplaceAll(data.EventData.GetStage(), " ", "-")
	labels = append(labels, "keptn_service:"+value)

	// Add result as a label (pass, warning or fail)
	labels = append(labels, "keptn_result:"+data.Evaluation.Result)

	for labelKey, labelValue := range data.Labels {
		// Replace spaces with dashes for the Key and Value
		labelKeyClean := strings.ReplaceAll(labelKey, " ", "-")
		labelValueClean := strings.ReplaceAll(labelValue, " ", "-")

		//Stick the cleaned key and value back together
		cleanKeyValueLabel := fmt.Sprint(labelKeyClean, ":", labelValueClean)

		labels = append(labels, cleanKeyValueLabel) // Append this "key":"value" using Sprint so as to not add spaces
	}

	return labels
}

func createZendeskTicketForEvaluationFinished(myKeptn *keptnv2.Keptn, data *keptnv2.EvaluationFinishedEventData) string {

	log.Println("[eventhandlers.go] Creating Zendesk Body details for evaluation.finished...")

	// Build title field (Zendesk ticket title)
	ticketTitle := "[EVALUATION] " + data.EventData.GetProject() + " - " + data.EventData.GetService() + " - " + data.EventData.GetStage() + " - Result: " + data.Evaluation.Result

	// Build description field (Zendesk ticket body)
	// Build result table
	// TODO - Zendesk doesn't accept Markdown by default. Reformat this...
	bodyContent := "||*Result*||*Score*||\n"

	result := ""
	if data.Evaluation.Result == "pass" {
		result += data.Evaluation.Result + " ✅"
	} else if data.Evaluation.Result == "warning" {
		result += data.Evaluation.Result + " ⚠"
	} else if data.Evaluation.Result == "fail" {
		result += data.Evaluation.Result + " ❌"
	} else {
		result += data.Evaluation.Result
	}
	bodyContent += "|" + result + "|" + fmt.Sprint(data.Evaluation.Score) + "|" + "\n\n"

	// Add Start Time and End Time
	bodyContent += "Start Time: " + data.Evaluation.TimeStart + "\n"
	bodyContent += "End Time: " + data.Evaluation.TimeEnd + "\n"

	// Add Keptn Context
	bodyContent += "Keptn Context ID: " + myKeptn.KeptnContext + "\n"

	// Add link to Keptn Bridge
	bridgeURL := KEPTN_DETAILS.BridgeURL + "/project/" + data.EventData.GetProject() + "/sequence/" + myKeptn.KeptnContext
	bodyContent += "[Link To Keptn's Bridge|" + bridgeURL + "]"

	// Build map of labels which we take from the cloudevent, which we then attach to the Zendesk ticket
	labels := createZendeskLabelsForEvaluationFinishedEvents(data)

	// Send the POST to Zendesk
	ticketKey := createZendeskTicket(ticketTitle, bodyContent, labels)

	return ticketKey
}

/**************************************
*         GENERIC METHODS
***************************************/
// Shared Function between evaluations and remediation finished events to create a Zendesk ticket
// By this point, summary and description are correctly formulated
// Depending on the type of ticket so this function can be shared
// As it just sends the POST to Zendesk
func createZendeskTicket(ticketTitle string, bodyContent string, labels []string) string {

	url := ZENDESK_DETAILS.BaseURL + "/api/v2/requests.json"
	jsonString, _ := json.Marshal(bodyContent)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonString))

	if err != nil {
		log.Println("[eventhandlers.go] Got an error message creating request: ", err)
	}

	username := ZENDESK_DETAILS.EndUserEmail + "/token"
	password := ZENDESK_DETAILS.APIToken
	req.SetBasicAuth(username, password)
	req.Header.Add("Content-Type", "application/json")
	req.Close = true

	client := http.Client{}

	// Send POST
	response, err := client.Do(req)
	if err != nil {
		log.Println("[eventhandlers.go] Got an error sending request", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Println("[eventhandlers.go] Got a non OK status code. Status is: ", response.Status)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("[eventhandlers.go] Got an error reading response body:", err)
	} else {
		log.Println("[eventhandlers.go] Response Body: ", body)
	}

	return ""

}
