// go-sentry-cli
//
// Simple application for creating http request to Sentry API over library
// go-sentry-api (more here: https://github.com/atlassian/go-sentry-api).
// Supported easy functional for check connection, create organization / project,
// print client keys of each project in the organization.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/atlassian/go-sentry-api"
)

var (
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

// Get info about organization if exists and return true, organization pointer
// or false, nil
func getOrganization(sentryClient *sentry.Client, orgSLug string) (bool, *sentry.Organization) {
	InfoLogger.Printf("Checking the organization <%s> for existence", orgSLug)

	org, err := sentryClient.GetOrganization(orgSLug)
	if err == nil {
		return true, &org
	}

	ErrorLogger.Println("Get organization info failed: ", err)

	return false, nil
}

// Create organization and return true, organization pointer if success
// else false, nil
func createOrganization(sentryClient *sentry.Client, orgSlug string) (bool, *sentry.Organization) {
	InfoLogger.Printf("Creating a non-existent organization <%s>", orgSlug)

	org, err := sentryClient.CreateOrganization(orgSlug)
	if err == nil {
		return true, &org
	}

	ErrorLogger.Printf("Creating organization <%s> failed: %s", orgSlug, err)

	return false, nil
}

func getProject(sentryClient *sentry.Client, org *sentry.Organization, projSlug string) (bool, *sentry.Project) {
	InfoLogger.Printf("Checking the existence of the project <%s> in the organization <%s> (id: %d)", projSlug, org.Name, org.ID)

	proj, err := sentryClient.GetProject(*org, projSlug)
	if err == nil {
		return true, &proj
	}

	ErrorLogger.Printf("Get <%s> project info from organization <%s> failed: %s", projSlug, *org.Slug, err)

	return false, nil
}

func createProject(sentryClient *sentry.Client, org *sentry.Organization, projSlug string) (bool, *sentry.Project) {
	InfoLogger.Printf("Creating a non-existent project <%s>", projSlug)

	team := sentry.Team{
		Slug:        org.Slug,
		Name:        org.Name,
	}

	proj, err := sentryClient.CreateProject(*org, team, projSlug, &projSlug)
	if err == nil {
		return true, &proj
	}

	ErrorLogger.Printf("Create <%s> project in organization <%s> failed: %s", projSlug, *org.Slug, err)

	return false, nil
}

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime)
}

func main() {
	var (
		host	 string
		apiToken string
		timeout	 int
		endpoint string
		tail	 []string
	)

	flag.StringVar(&host, "host", "", "* Host of your Sentry instance including the protocol \n" +
		"  Example: https://sentry.io")
	flag.StringVar(&apiToken, "token", "", "* Personal access token generated by your Sentry instance")
	flag.IntVar(&timeout, "timeout", 10, "Connection timeout during a HTTP request")
	flag.Parse()

	if host == "" || apiToken == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	endpoint = fmt.Sprintf("%s/api/0/", host)
	sentryClient, _ := sentry.NewClient(apiToken, &endpoint, &timeout)

	InfoLogger.Println("Check connection with provided parameters")

	if _, _, err := sentryClient.GetOrganizations(); err != nil {
		ErrorLogger.Println("Connection failed: ", err)
		os.Exit(2)
	}

	InfoLogger.Println("Connection success!")

	if len(flag.Args()) < 2 {
		InfoLogger.Println("No more parameters were passed. No more actions are required")
		os.Exit(0)
	}

	tail = flag.Args()
	orgSlug := strings.ToLower(tail[0])
	projSlug := strings.ToLower(tail[1])

	eO, org := getOrganization(sentryClient, orgSlug)
	if !eO {
		_, org = createOrganization(sentryClient, orgSlug)
	}

	eP, proj := getProject(sentryClient, org, projSlug)
	if !eP {
		_, proj = createProject(sentryClient, org, projSlug)
	}

	clientKeys, _ := sentryClient.GetClientKeys(*org, *proj)

	for _, key := range clientKeys {
		InfoLogger.Printf("Your public DSN: %s", key.DSN.Public)
	}
}
