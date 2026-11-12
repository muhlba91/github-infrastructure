package google_test

import (
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/muhlba91/github-infrastructure/pkg/lib/google"
	googleConf "github.com/muhlba91/github-infrastructure/pkg/model/config/google"
	googleModel "github.com/muhlba91/github-infrastructure/pkg/model/google"
	"github.com/pulumi/pulumi-gcp/sdk/v10/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceMocks records the created gcp:projects/service:Service resources as "<project>/<service>".
type serviceMocks struct {
	mu       sync.Mutex
	services []string
}

func (m *serviceMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if args.TypeToken == "gcp:projects/service:Service" {
		m.mu.Lock()
		m.services = append(m.services,
			args.Inputs["project"].StringValue()+"/"+args.Inputs["service"].StringValue())
		m.mu.Unlock()
	}
	return args.Name + "_id", args.Inputs, nil
}

func (m *serviceMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func testRepositoryProjects() map[string]*googleModel.RepositoryProject {
	project := "main-project"
	return map[string]*googleModel.RepositoryProject{
		"repo": {
			Repository:      new("repo"),
			Name:            &project,
			EnabledServices: slices.Concat([]string{"dns.googleapis.com"}, google.DefaultServices()),
			LinkedProjects: map[string]*googleModel.RepositoryLinkedProject{
				"linked-default": {AccessLevel: "default"},
				"linked-full":    {AccessLevel: "full"},
			},
		},
	}
}

func TestProjectServices(t *testing.T) {
	repoProject := testRepositoryProjects()["repo"]

	tests := []struct {
		project string
		want    []string
	}{
		{"main-project", repoProject.EnabledServices},
		{"linked-full", repoProject.EnabledServices},
		{"linked-default", google.DefaultServices()},
		{"unrelated", nil},
	}
	for _, tt := range tests {
		t.Run(tt.project, func(t *testing.T) {
			// a copy of the name, so the comparison must not depend on the pointer
			project := tt.project
			got := google.ProjectServices(repoProject, project)
			if !slices.Equal(got, tt.want) {
				t.Errorf("projectServices(%q) = %v, want %v", project, got, tt.want)
			}
		})
	}
}

func TestDefaultServicesContainSecretManager(t *testing.T) {
	if !slices.Contains(google.DefaultServices(), google.SecretManagerService) {
		t.Errorf("google.DefaultServices() must contain %s", google.SecretManagerService)
	}
}

func TestEnableProjectServices(t *testing.T) {
	mocks := &serviceMocks{}
	gcpConfig := &googleConf.Config{
		Projects: []string{"main-project", "linked-default", "linked-full", "unrelated"},
	}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		providers := map[string]*gcp.Provider{}
		for _, project := range gcpConfig.Projects {
			provider, pErr := gcp.NewProvider(ctx, "provider-"+project, &gcp.ProviderArgs{})
			if pErr != nil {
				return pErr
			}
			providers[project] = provider
		}

		enabled, eErr := google.EnableProjectServices(ctx, testRepositoryProjects(), gcpConfig, providers)
		if eErr != nil {
			return eErr
		}
		if len(enabled["unrelated"]) != 0 {
			t.Errorf("unrelated project: got %d services, want 0", len(enabled["unrelated"]))
		}
		return nil
	}, pulumi.WithMocks("project", "stack", mocks))
	if err != nil {
		t.Fatalf("EnableProjectServices: %v", err)
	}

	for _, want := range []string{
		"main-project/dns.googleapis.com",
		"main-project/" + google.SecretManagerService,
		"linked-full/dns.googleapis.com",
		"linked-default/" + google.SecretManagerService,
	} {
		if !slices.Contains(mocks.services, want) {
			t.Errorf("service %s not enabled", want)
		}
	}
	if slices.Contains(mocks.services, "linked-default/dns.googleapis.com") {
		t.Error("linked project with access level default must not get the repository's services")
	}
	for _, svc := range mocks.services {
		if strings.HasPrefix(svc, "unrelated/") {
			t.Errorf("unrelated project got service %s", svc)
		}
	}
	// main-project and linked-full: the defaults and dns.googleapis.com; linked-default: the defaults
	defaults := len(google.DefaultServices())
	if got, want := len(mocks.services), 2*(defaults+1)+defaults; got != want {
		t.Errorf("got %d services, want %d", got, want)
	}
}
