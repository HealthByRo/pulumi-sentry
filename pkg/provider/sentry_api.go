package provider

import "github.com/atlassian/go-sentry-api"

// sentryClientAPI is an interface that covers all the functionality we need
// from sentry.Client.
type sentryClientAPI interface {
	CreateProject(o sentry.Organization, t sentry.Team, name string, slug *string) (sentry.Project, error)
	GetProject(o sentry.Organization, projslug string) (sentry.Project, error)
	UpdateProject(o sentry.Organization, p sentry.Project) error
	DeleteProject(o sentry.Organization, p sentry.Project) error

	CreateClientKey(o sentry.Organization, p sentry.Project, name string) (sentry.Key, error)
	DeleteClientKey(o sentry.Organization, p sentry.Project, k sentry.Key) error
	UpdateClientKey(o sentry.Organization, p sentry.Project, k sentry.Key, name string) (sentry.Key, error)
	GetClientKeys(o sentry.Organization, p sentry.Project) ([]sentry.Key, error)

	GetOrganization(orgslug string) (sentry.Organization, error)

	GetTeam(o sentry.Organization, teamSlug string) (sentry.Team, error)
}

// sentryClientMock mocks sentry.Client for tests.
type sentryClientMock struct {
	sentryClientAPI

	createProject func(o sentry.Organization, t sentry.Team, name string, slug *string) (sentry.Project, error)
	getProject    func(o sentry.Organization, projslug string) (sentry.Project, error)
	updateProject func(o sentry.Organization, p sentry.Project) error
	deleteProject func(o sentry.Organization, p sentry.Project) error

	createClientKey func(o sentry.Organization, p sentry.Project, name string) (sentry.Key, error)
	deleteClientKey func(o sentry.Organization, p sentry.Project, k sentry.Key) error
	updateClientKey func(o sentry.Organization, p sentry.Project, k sentry.Key, name string) (sentry.Key, error)
	getClientKeys   func(o sentry.Organization, p sentry.Project) ([]sentry.Key, error)

	getOrganization func(orgslug string) (sentry.Organization, error)

	getTeam func(o sentry.Organization, teamSlug string) (sentry.Team, error)
}

func (m *sentryClientMock) CreateProject(o sentry.Organization, t sentry.Team, name string, slug *string) (sentry.Project, error) {
	return m.createProject(o, t, name, slug)
}

func (m *sentryClientMock) GetProject(o sentry.Organization, projslug string) (sentry.Project, error) {
	return m.getProject(o, projslug)
}

func (m *sentryClientMock) UpdateProject(o sentry.Organization, p sentry.Project) error {
	return m.updateProject(o, p)
}

func (m *sentryClientMock) DeleteProject(o sentry.Organization, p sentry.Project) error {
	return m.deleteProject(o, p)
}

func (m *sentryClientMock) CreateClientKey(o sentry.Organization, p sentry.Project, name string) (sentry.Key, error) {
	return m.createClientKey(o, p, name)
}

func (m *sentryClientMock) DeleteClientKey(o sentry.Organization, p sentry.Project, k sentry.Key) error {
	return m.DeleteClientKey(o, p, k)
}

func (m *sentryClientMock) UpdateClientKey(o sentry.Organization, p sentry.Project, k sentry.Key, name string) (sentry.Key, error) {
	return m.UpdateClientKey(o, p, k, name)
}

func (m *sentryClientMock) GetClientKeys(o sentry.Organization, p sentry.Project) ([]sentry.Key, error) {
	return m.GetClientKeys(o, p)
}

func (m *sentryClientMock) GetOrganization(orgslug string) (sentry.Organization, error) {
	return m.getOrganization(orgslug)
}

func (m *sentryClientMock) GetTeam(o sentry.Organization, teamSlug string) (sentry.Team, error) {
	return m.getTeam(o, teamSlug)
}
