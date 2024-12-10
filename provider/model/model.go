//go:generate go run ../../pkg/sdk/runable/steampipe_es_client_generator/main.go -pluginPath ../../steampipe-plugin-REPLACEME/REPLACEME -file $GOFILE -output ../../pkg/sdk/es/resources_clients.go -resourceTypesFile ../resource_types/resource-types.json

// Implement types for each resource

package model

import "time"

type Metadata struct{}

type Owner struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	TwoFactorAuthEnabled bool   `json:"twoFactorAuthEnabled"`
	Type                 string `json:"type"`
}

type ProjectResponse struct {
	Project ProjectDescription `json:"project"`
	Cursor  string             `json:"cursor"`
}

type ProjectDescription struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Name           string    `json:"name"`
	Owner          Owner     `json:"owner"`
	EnvironmentIDs []string  `json:"environmentIds"`
}

type EnvironmentResponse struct {
	Environment EnvironmentDescription `json:"environment"`
	Cursor      string                 `json:"cursor"`
}

type EnvironmentDescription struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	ProjectID       string   `json:"projectId"`
	DatabasesIDs    []string `json:"databasesIds"`
	RedisIDs        []string `json:"redisIds"`
	ServiceIDs      []string `json:"serviceIds"`
	EnvGroupIDs     []string `json:"envGroupIds"`
	ProtectedStatus string   `json:"protectedStatus"`
}

type IPAllow struct {
	CIDRBlock   string `json:"cidrBlock"`
	Description string `json:"description"`
}

type ReadReplica struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PostgresResponse struct {
	Postgres PostgresDescription `json:"postgres"`
	Cursor   string              `json:"cursor"`
}

type PostgresDescription struct {
	ID                      string        `json:"id"`
	IPAllowList             []IPAllow     `json:"ipAllowList"`
	CreatedAt               time.Time     `json:"createdAt"`
	UpdatedAt               time.Time     `json:"updatedAt"`
	ExpiresAt               time.Time     `json:"expiresAt"`
	DatabaseName            string        `json:"databaseName"`
	DatabaseUser            string        `json:"databaseUser"`
	EnvironmentID           string        `json:"environmentId"`
	HighAvailabilityEnabled bool          `json:"highAvailabilityEnabled"`
	Name                    string        `json:"name"`
	Owner                   Owner         `json:"owner"`
	Plan                    string        `json:"plan"`
	DiskSizeGB              int           `json:"diskSizeGB"`
	PrimaryPostgresID       string        `json:"primaryPostgresID"`
	Region                  string        `json:"region"`
	ReadReplicas            []ReadReplica `json:"readReplicas"`
	Role                    string        `json:"role"`
	Status                  string        `json:"status"`
	Version                 string        `json:"version"`
	Suspended               string        `json:"suspended"`
	Suspenders              []string      `json:"suspenders"`
	DashboardURL            string        `json:"dashboardUrl"`
}

type BuildFilter struct {
	Paths        []string `json:"paths"`
	IgnoredPaths []string `json:"ignoredPaths"`
}

type RegistryCredential struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ServiceDetails struct {
	BuildCommand string       `json:"buildCommand"`
	ParentServer ParentServer `json:"parentServer"`
	PublishPath  string       `json:"publishPath"`
	Previews     Previews     `json:"previews"`
	URL          string       `json:"url"`
	BuildPlan    string       `json:"buildPlan"`
}

type ParentServer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Previews struct {
	Generation string `json:"generation"`
}

type ServiceResponse struct {
	Service ServiceDescription `json:"service"`
	Cursor  string             `json:"cursor"`
}

type ServiceDescription struct {
	ID                 string             `json:"id"`
	AutoDeploy         string             `json:"autoDeploy"`
	Branch             string             `json:"branch"`
	BuildFilter        BuildFilter        `json:"buildFilter"`
	CreatedAt          time.Time          `json:"createdAt"`
	DashboardURL       string             `json:"dashboardUrl"`
	EnvironmentID      string             `json:"environmentId"`
	ImagePath          string             `json:"imagePath"`
	Name               string             `json:"name"`
	NotifyOnFail       string             `json:"notifyOnFail"`
	OwnerID            string             `json:"ownerId"`
	RegistryCredential RegistryCredential `json:"registryCredential"`
	Repo               string             `json:"repo"`
	RootDir            string             `json:"rootDir"`
	Slug               string             `json:"slug"`
	Suspended          string             `json:"suspended"`
	Suspenders         []string           `json:"suspenders"`
	Type               string             `json:"type"`
	UpdatedAt          time.Time          `json:"updatedAt"`
	ServiceDetails     ServiceDetails     `json:"serviceDetails"`
}

type JobResponse struct {
	Job    JobDescription `json:"job"`
	Cursor string         `json:"cursor"`
}

type JobDescription struct {
	ID           string    `json:"id"`
	ServiceID    string    `json:"serviceId"`
	StartCommand string    `json:"startCommand"`
	PlanID       string    `json:"planId"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
}

type DiskResponse struct {
	Disk   DiskDescription `json:"disk"`
	Cursor string          `json:"cursor"`
}

type DiskDescription struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SizeGB    int       `json:"sizeGB"`
	MountPath string    `json:"mountPath"`
	ServiceID string    `json:"serviceId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Commit struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type Image struct {
	Ref                string `json:"ref"`
	SHA                string `json:"sha"`
	RegistryCredential string `json:"registryCredential"`
}

type DeployResponse struct {
	Deploy DeployDescription `json:"deploy"`
	Cursor string            `json:"cursor"`
}

type DeployDescription struct {
	ID         string    `json:"id"`
	Commit     Commit    `json:"commit"`
	Image      Image     `json:"image"`
	Status     string    `json:"status"`
	Trigger    string    `json:"trigger"`
	FinishedAt time.Time `json:"finishedAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type BlueprintResponse struct {
	BluePrint BlueprintDescription `json:"blueprint"`
	Cursor    string               `json:"cursor"`
}

type BlueprintDescription struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	AutoSync bool      `json:"autoSync"`
	Repo     string    `json:"repo"`
	Branch   string    `json:"branch"`
	LastSync time.Time `json:"lastSync"`
}

type ServiceLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type EnvGroupDescription []struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	OwnerID       string        `json:"ownerId"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
	ServiceLinks  []ServiceLink `json:"serviceLinks"`
	EnvironmentID string        `json:"environmentId"`
}

type HeaderResponse struct {
	Header HeaderDescription `json:"header"`
	Cursor string            `json:"cursor"`
}

type HeaderDescription struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RouteResponse struct {
	Route  RouteDescription `json:"route"`
	Cursor string           `json:"cursor"`
}

type RouteDescription struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Priority    int    `json:"priority"`
}
