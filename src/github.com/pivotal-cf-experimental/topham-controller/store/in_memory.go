package store

import (
	"fmt"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type ServiceInstance struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	PlanID      string `json:"plan_id"`
	PlanName    string `json:"plan_name"`
}

type AnnotatedBinding struct {
	ID      string           `json:"id"`
	Binding osb.BindResponse `json:"binding"`
}

type Store struct {
	ServiceInstances map[string]ServiceInstance
	Bindings         map[string]map[string]osb.BindResponse
	Catalog          osb.CatalogResponse
}

func NewStore(catalog osb.CatalogResponse) *Store {
	serviceInstances := make(map[string]ServiceInstance)
	namedBindings := make(map[string]map[string]osb.BindResponse)

	return &Store{
		ServiceInstances: serviceInstances,
		Catalog:          catalog,
		Bindings:         namedBindings,
	}
}

func (s *Store) ListServiceInstances() []ServiceInstance {
	instanceList := []ServiceInstance{}
	for _, v := range s.ServiceInstances {
		instanceList = append(instanceList, v)
	}

	return instanceList
}

//TODO handle missing case
func (s *Store) GetServiceInstance(name string) (instance ServiceInstance, found bool, err error) {
	instance, found = s.ServiceInstances[name]
	return instance, found, nil
}

func (s *Store) CreateServiceInstance(name, serviceID, planID string) error {
	serviceInst := ServiceInstance{
		ID:          name,
		Name:        name,
		ServiceID:   serviceID,
		ServiceName: s.getServiceNameForID(serviceID),
		PlanID:      planID,
		PlanName:    s.getPlanNameForID(planID),
	}

	if _, ok := s.ServiceInstances[name]; ok {
		return fmt.Errorf("service instance %s already exists", name)
	}

	s.ServiceInstances[name] = serviceInst
	return nil
}

func (s *Store) DeleteServiceInstance(name string) error {
	if _, ok := s.ServiceInstances[name]; !ok {
		return fmt.Errorf("service instance %s does not exist", name)
	}
	fmt.Println("deleteing name")

	delete(s.ServiceInstances, name)
	return nil
}

func (s *Store) GetCatalog() osb.CatalogResponse {
	return s.Catalog
}

func (s *Store) CreateBinding(instanceID, bindingID string, binding osb.BindResponse) error {
	_, found := s.Bindings[instanceID]
	if !found {
		s.Bindings[instanceID] = make(map[string]osb.BindResponse)
	}

	instanceBindings, _ := s.Bindings[instanceID]
	if _, found = instanceBindings[bindingID]; found {
		return fmt.Errorf("binding %s already exists for instance %s", bindingID, instanceID)
	}

	instanceBindings[bindingID] = binding
	return nil
}

func (s *Store) ListBindingsForInstance(instanceID string) ([]AnnotatedBinding, error) {
	var returnedBindings []AnnotatedBinding

	instanceBindings, found := s.Bindings[instanceID]
	if !found {
		return nil, fmt.Errorf("no bindings found for instance %s", instanceID)
	}

	for i, binding := range instanceBindings {
		returnedBindings = append(returnedBindings, AnnotatedBinding{
			ID:      i,
			Binding: binding,
		})
	}

	return returnedBindings, nil
}

func (s *Store) DeleteBinding(instanceID, bindingID string) error {
	instanceBindings, found := s.Bindings[instanceID]
	if !found {
		return fmt.Errorf("no bindings found for instance %s", instanceID)
	}

	if _, found := instanceBindings[bindingID]; !found {
		return fmt.Errorf("binding %s not found for instance %s", bindingID, instanceID)
	}

	delete(instanceBindings, bindingID)
	return nil
}

func (s *Store) getServiceNameForID(id string) string {
	for _, v := range s.Catalog.Services {
		if v.ID == id {
			return v.Name
		}
	}
	return ""
}

func (s *Store) getPlanNameForID(planID string) string {
	for _, service := range s.Catalog.Services {
		for _, plan := range service.Plans {
			if plan.ID == planID {
				return plan.Name
			}
		}
	}
	return ""
}
