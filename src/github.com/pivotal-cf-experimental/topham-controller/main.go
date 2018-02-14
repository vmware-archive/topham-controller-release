package main

import (
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/topham-controller/api"
	"github.com/pivotal-cf-experimental/topham-controller/store"
	"github.com/pivotal-cf/brokerapi"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

var brokerURL = os.Getenv("BROKER_URL")
var username = os.Getenv("BROKER_USERNAME")
var password = os.Getenv("BROKER_PASSWORD")

var instancesStore *store.Store

func main() {
	brokerClient := createBrokerClient()

	catalog, err := brokerClient.GetCatalog()
	if err != nil {
		log.Fatal(err)
	}

	instancesStore = store.NewStore(*catalog)

	ctrl := api.NewServicesController(brokerClient, instancesStore)

	r := mux.NewRouter()
	logger := lager.NewLogger("ServicesController")
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.DEBUG))

	brokerapi.AttachRoutes(r, ctrl, logger)
	r.HandleFunc("/v2/service_instances", ctrl.ListInstancesHandler).Methods("GET")
	//TODO how does this relate to this proposal
	//https://github.com/mattmcneeney/servicebroker/blob/9a47c7cac98d2145ffcf9ef5049863113d1f8af7/spec.md#fetching-a-service-instance
	r.HandleFunc("/v2/service_instances/{name}", ctrl.GetInstanceHandler).Methods("GET")
	r.HandleFunc("/v2/service_instances/{name}/service_bindings", ctrl.ListBindingsHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createBrokerClient() osb.Client {
	config := osb.DefaultClientConfiguration()
	config.URL = brokerURL
	config.AuthConfig = &osb.AuthConfig{
		BasicAuthConfig: &osb.BasicAuthConfig{
			Username: username,
			Password: password,
		},
	}

	client, err := osb.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	return client
}
