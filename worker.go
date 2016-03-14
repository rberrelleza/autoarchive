package main

import (
	"github.com/tbruyelle/hipchat-go/hipchat"
)

func StartWorker() {
	b := NewBackendServer("hiparchiver.workers")

	taskServer := NewTaskServer()
	taskServer.RegisterTask("autoArchive", b.AutoArchive)

	worker := taskServer.NewWorker("worker1")
	err := worker.Launch()
	if err != nil {
  	panic(err)
	}
}

// Machinery requires to return a (interface{}, error) even if we don't handle
// the result, so faking it for now (shrug)
func (s *Server) AutoArchive(tenantID string) (bool, error) {
	w := Worker{
		Log:         s.Log,
	}

	tenants := s.NewTenants()
	tenant, err := tenants.Get(tenantID)
	if err != nil {
		s.Log.Errorf("Coudn't find tid-%s", tenantID)
		return true, err
	}

	tenantConfigurations := s.NewTenantConfigurations()
	tenantConfiguration, err := tenantConfigurations.Get(tenantID)

	if err != nil {
		s.Log.Errorf("Coudn't find a configuration for tid-%s", tenantID)
		return true, err
	}

	credentials := hipchat.ClientCredentials{
		ClientID:     tenant.ID,
		ClientSecret: tenant.Secret,
	}

	newClient := hipchat.NewClient("")
	token, _, err := newClient.GenerateToken(
		credentials,
		[]string{hipchat.ScopeManageRooms, hipchat.ScopeViewGroup, hipchat.ScopeSendNotification, hipchat.ScopeAdminRoom})

	if err != nil {
		// this typically means the group uninstalled the plugin
		s.Log.Errorf("Client.GetAccessToken returned an error %v", err)
		return true, err
	}

	client := token.CreateClient()
	rooms, error := w.GetRooms(client)
	if error != nil {
		s.Log.Errorf("Failed to retrieve rooms for tid-%d", tenantID)
		return true, error
	}

	for _, room := range rooms {
		w.MaybeArchiveRoom(tenantID, room.ID, tenantConfiguration.Threshold, client)
	}

	s.Log.Infof("Finished processing rooms for tid-%s", tenantID)
	return true, nil
}
