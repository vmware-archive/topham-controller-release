package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/topham-controller/store"
	"github.com/pivotal-cf/brokerapi"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type Store interface {
	GetCatalog() osb.CatalogResponse
	CreateServiceInstance(name, planID, serviceID string) error
	DeleteServiceInstance(name string) error
	GetServiceInstance(name string) (store.ServiceInstance, bool, error)
	ListServiceInstances() []store.ServiceInstance
	CreateBinding(instanceID, bindingID string, binding osb.BindResponse) error
	ListBindingsForInstance(instanceID string) ([]store.AnnotatedBinding, error)
	DeleteBinding(instanceID, bindingID string) error
}

type ServicesController struct {
	client osb.Client
	store  Store
}

func NewServicesController(client osb.Client, store Store) *ServicesController {
	return &ServicesController{
		client: client,
		store:  store,
	}
}

func (s *ServicesController) Services(ctx context.Context) []brokerapi.Service {
	servicesRaw, err := json.Marshal(s.store.GetCatalog().Services)
	if err != nil {
		//TODO handle
		log.Println("ERROR: " + err.Error())
	}

	var services []brokerapi.Service

	err = json.Unmarshal(servicesRaw, &services)
	if err != nil {
		//TODO handle
		log.Println("ERROR: " + err.Error())
	}

	return services
}

func (s *ServicesController) Unbind(ctx context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	preq := osb.UnbindRequest{
		InstanceID:          instanceID,
		BindingID:           bindingID,
		AcceptsIncomplete:   false,
		ServiceID:           details.ServiceID,
		PlanID:              details.PlanID,
		OriginatingIdentity: nil,
	}

	_, err := s.client.Unbind(&preq)

	if err != nil {
		fmt.Printf(err.Error())
		return err
	}

	err = s.store.DeleteBinding(instanceID, bindingID)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServicesController) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	var resp brokerapi.ProvisionedServiceSpec

	preq := osb.ProvisionRequest{
		InstanceID:        instanceID,
		ServiceID:         details.ServiceID,
		PlanID:            details.PlanID,
		OrganizationGUID:  "dummy-org-id",
		SpaceGUID:         "dummy-space-id",
		AcceptsIncomplete: true,
	}

	_, found, _ := s.store.GetServiceInstance(instanceID)
	if found {
		return resp, fmt.Errorf("Instance %s already exists", instanceID)
	}

	provisionResponse, err := s.client.ProvisionInstance(&preq)
	if err != nil {
		return resp, err
	}

	err = s.store.CreateServiceInstance(instanceID, details.ServiceID, details.PlanID)
	if err != nil {
		//TODO orphan mitigation
		return resp, err
	}

	resp.IsAsync = provisionResponse.Async

	if provisionResponse.DashboardURL != nil {
		resp.DashboardURL = *provisionResponse.DashboardURL
	}

	if provisionResponse.OperationKey != nil {
		resp.OperationData = string(*provisionResponse.OperationKey)
	}

	return resp, nil
}

func (s *ServicesController) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	var resp brokerapi.DeprovisionServiceSpec

	preq := osb.DeprovisionRequest{
		InstanceID:        instanceID,
		ServiceID:         details.ServiceID,
		PlanID:            details.PlanID,
		AcceptsIncomplete: true,
	}

	//TODO Does the instance exist?
	brokerResp, err := s.client.DeprovisionInstance(&preq)
	if err != nil {
		return resp, err
	}

	err = s.store.DeleteServiceInstance(instanceID)
	if err != nil {
		//TODO orphan mitigation
		return resp, err
	}

	resp.IsAsync = brokerResp.Async

	if brokerResp.OperationKey != nil {
		resp.OperationData = string(*brokerResp.OperationKey)
	}

	return resp, nil
}

func (s *ServicesController) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	var bindingResponse brokerapi.Binding
	preq := osb.BindRequest{
		BindingID:         bindingID,
		InstanceID:        instanceID,
		AcceptsIncomplete: false,
		ServiceID:         details.ServiceID,
		PlanID:            details.PlanID,
	}

	brokerResp, err := s.client.Bind(&preq)
	if err != nil {
		return bindingResponse, err
	}

	err = s.store.CreateBinding(instanceID, bindingID, *brokerResp)
	if err != nil {
		//TODO orphan mitigation
		return bindingResponse, err
	}

	bytes, err := json.Marshal(brokerResp)
	if err != nil {
		return bindingResponse, err
	}

	err = json.Unmarshal(bytes, &bindingResponse)
	if err != nil {
		return bindingResponse, err
	}

	return bindingResponse, nil
}

func (s *ServicesController) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, nil
}

func (s *ServicesController) LastOperation(ctx context.Context, instanceID, operationData string) (brokerapi.LastOperation, error) {
	var resp brokerapi.LastOperation
	req := osb.LastOperationRequest{
		InstanceID: instanceID,
	}

	brokerResp, err := s.client.PollLastOperation(&req)
	if err != nil {
		return resp, err
	}

	resp.State = brokerapi.LastOperationState(brokerResp.State)

	if brokerResp.Description != nil {
		resp.Description = *brokerResp.Description
	}
	return resp, nil
}

func (s *ServicesController) ListBindingsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceName := vars["name"]
	bindings, err := s.store.ListBindingsForInstance(instanceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	bytes, err := json.Marshal(bindings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(bytes)
}

func (s *ServicesController) ListInstancesHandler(w http.ResponseWriter, r *http.Request) {
	instances := s.store.ListServiceInstances()
	bytes, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(bytes)
}

func (s *ServicesController) GetInstanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceName := vars["name"]

	instance, found, err := s.store.GetServiceInstance(instanceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if !found {
		http.Error(w, "instance not found", http.StatusNotFound)
	}

	bytes, err := json.Marshal(instance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(bytes)
}
