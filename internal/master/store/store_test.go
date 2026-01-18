package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSatellites(t *testing.T) {
	s := testStore(t)
	defer s.Close()

	sat := &Satellite{
		ID:       "sat-001",
		Hostname: "test-host",
		OS:       "linux",
		Arch:     "amd64",
		Status:   SatelliteStatusOnline,
	}

	if err := s.RegisterSatellite(sat); err != nil {
		t.Fatalf("RegisterSatellite failed: %v", err)
	}

	got, err := s.GetSatellite("sat-001")
	if err != nil {
		t.Fatalf("GetSatellite failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetSatellite returned nil")
	}
	if got.Hostname != "test-host" {
		t.Errorf("Hostname = %s, want test-host", got.Hostname)
	}

	sats, err := s.ListSatellites()
	if err != nil {
		t.Fatalf("ListSatellites failed: %v", err)
	}
	if len(sats) != 1 {
		t.Errorf("ListSatellites returned %d, want 1", len(sats))
	}

	if err := s.UpdateSatelliteStatus("sat-001", SatelliteStatusOffline); err != nil {
		t.Fatalf("UpdateSatelliteStatus failed: %v", err)
	}
	got, _ = s.GetSatellite("sat-001")
	if got.Status != SatelliteStatusOffline {
		t.Errorf("Status = %s, want offline", got.Status)
	}

	if err := s.DeleteSatellite("sat-001"); err != nil {
		t.Fatalf("DeleteSatellite failed: %v", err)
	}
	got, _ = s.GetSatellite("sat-001")
	if got != nil {
		t.Error("After delete, satellite still exists")
	}
}

func TestModules(t *testing.T) {
	s := testStore(t)
	defer s.Close()

	mod := &Module{
		ID:      "ssh-honeypot",
		Name:    "SSH Honeypot",
		Version: "1.0.0",
		Digest:  "sha256:abc123",
	}

	if err := s.AddModule(mod); err != nil {
		t.Fatalf("AddModule failed: %v", err)
	}

	got, err := s.GetModule("ssh-honeypot", "1.0.0")
	if err != nil {
		t.Fatalf("GetModule failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetModule returned nil")
	}
	if got.Name != "SSH Honeypot" {
		t.Errorf("Name = %s, want SSH Honeypot", got.Name)
	}

	mod2 := &Module{
		ID:      "ssh-honeypot",
		Name:    "SSH Honeypot",
		Version: "2.0.0",
		Digest:  "sha256:def456",
	}
	s.AddModule(mod2)

	versions, err := s.ListModuleVersions("ssh-honeypot")
	if err != nil {
		t.Fatalf("ListModuleVersions failed: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("ListModuleVersions returned %d, want 2", len(versions))
	}

	latest, err := s.GetLatestModule("ssh-honeypot")
	if err != nil {
		t.Fatalf("GetLatestModule failed: %v", err)
	}
	if latest.Version != "2.0.0" {
		t.Errorf("Latest version = %s, want 2.0.0", latest.Version)
	}

	if err := s.DeleteModule("ssh-honeypot", "1.0.0"); err != nil {
		t.Fatalf("DeleteModule failed: %v", err)
	}
	versions, _ = s.ListModuleVersions("ssh-honeypot")
	if len(versions) != 1 {
		t.Errorf("After delete, versions = %d, want 1", len(versions))
	}
}

func TestDeployments(t *testing.T) {
	s := testStore(t)
	defer s.Close()

	dep := &Deployment{
		SatelliteID: "sat-001",
		Modules: []ModuleDeployment{
			{
				ModuleID:      "ssh-honeypot",
				ModuleVersion: "1.0.0",
				Enabled:       true,
			},
		},
	}

	if err := s.SetDeployment(dep); err != nil {
		t.Fatalf("SetDeployment failed: %v", err)
	}

	got, err := s.GetDeployment("sat-001")
	if err != nil {
		t.Fatalf("GetDeployment failed: %v", err)
	}
	if got == nil {
		t.Fatal("GetDeployment returned nil")
	}
	if len(got.Modules) != 1 {
		t.Errorf("Modules count = %d, want 1", len(got.Modules))
	}

	if err := s.AddModuleToDeployment("sat-001", ModuleDeployment{
		ModuleID:      "http-honeypot",
		ModuleVersion: "1.0.0",
	}); err != nil {
		t.Fatalf("AddModuleToDeployment failed: %v", err)
	}

	got, _ = s.GetDeployment("sat-001")
	if len(got.Modules) != 2 {
		t.Errorf("After add, modules = %d, want 2", len(got.Modules))
	}

	if err := s.RemoveModuleFromDeployment("sat-001", "ssh-honeypot"); err != nil {
		t.Fatalf("RemoveModuleFromDeployment failed: %v", err)
	}
	got, _ = s.GetDeployment("sat-001")
	if len(got.Modules) != 1 {
		t.Errorf("After remove, modules = %d, want 1", len(got.Modules))
	}

	s.SetDeployment(&Deployment{
		SatelliteID: "sat-002",
		Modules:     []ModuleDeployment{{ModuleID: "http-honeypot"}},
	})
	sats, _ := s.GetSatellitesByModule("http-honeypot")
	if len(sats) != 2 {
		t.Errorf("GetSatellitesByModule = %d, want 2", len(sats))
	}
}

func TestStaleSatellites(t *testing.T) {
	s := testStore(t)
	defer s.Close()

	sat := &Satellite{
		ID:         "sat-old",
		LastSeenAt: time.Now().Add(-10 * time.Minute),
	}
	s.db.PutJSON(BucketSatellites, sat.ID, sat)

	sat2 := &Satellite{
		ID:         "sat-new",
		LastSeenAt: time.Now(),
	}
	s.RegisterSatellite(sat2)

	stale, err := s.GetStaleSatellites(5 * time.Minute)
	if err != nil {
		t.Fatalf("GetStaleSatellites failed: %v", err)
	}
	if len(stale) != 1 {
		t.Errorf("Stale count = %d, want 1", len(stale))
	}
	if stale[0].ID != "sat-old" {
		t.Errorf("Stale ID = %s, want sat-old", stale[0].ID)
	}
}

func testStore(t *testing.T) *Store {
	t.Helper()
	tmpDir := t.TempDir()

	s, err := New(&Config{
		DBPath:   filepath.Join(tmpDir, "test.db"),
		ImageDir: filepath.Join(tmpDir, "images"),
	})
	if err != nil {
		t.Fatalf("New store failed: %v", err)
	}
	return s
}
