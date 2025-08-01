package scim

import (
	"fmt"
	"net/url"

	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// SCIMEndpoint is the base endpoint for SCIM API
	SCIMEndpoint = "/scim/v1/organizations"
)

// UserName represents the name structure for a user
type UserName struct {
	GivenName  string `json:"givenName,omitempty"`  // First name, maxLength: 5000
	FamilyName string `json:"familyName,omitempty"` // Last name, maxLength: 5000
}

// UserEmail represents an email structure for a user
type UserEmail struct {
	Value   string `json:"value,omitempty"`   // Email address, maxLength: 5000
	Primary bool   `json:"primary,omitempty"` // Whether this is the primary email
}

// UserRole represents a role structure for a user
type UserRole struct {
	Value   string `json:"value,omitempty"`   // Role value (e.g., "employee", "owner"), maxLength: 5000
	Primary bool   `json:"primary,omitempty"` // Whether this is the primary role
	Display string `json:"display,omitempty"` // Display name for the role
}

// UserMeta represents metadata about a user resource
type UserMeta struct {
	ResourceType string `json:"resourceType,omitempty"` // Type of resource (always "User")
	Created      string `json:"created,omitempty"`      // ISO 8601 creation timestamp
	LastModified string `json:"lastModified,omitempty"` // ISO 8601 last modification timestamp
	Location     string `json:"location,omitempty"`     // URL of the user resource
}

// User represents a SCIM user resource
type User struct {
	Schemas    []string    `json:"schemas,omitempty"`    // SCIM schemas used
	ID         string      `json:"id,omitempty"`         // Unique identifier
	ExternalID string      `json:"externalId,omitempty"` // External identifier from SAML provider, maxLength: 5000
	UserName   string      `json:"userName,omitempty"`   // Email address as username, maxLength: 5000
	Name       *UserName   `json:"name,omitempty"`       // User's name
	Emails     []UserEmail `json:"emails,omitempty"`     // List of user emails
	Roles      []UserRole  `json:"roles,omitempty"`      // List of user roles
	Active     bool        `json:"active"`               // Whether the user is active
	Meta       *UserMeta   `json:"meta,omitempty"`       // Resource metadata
}

// CreateUserParams represents parameters for creating a user
type CreateUserParams struct {
	UserName   string      `json:"userName"`             // Required: Email address as username, maxLength: 5000
	Name       *UserName   `json:"name,omitempty"`       // User's name
	Emails     []UserEmail `json:"emails,omitempty"`     // List of user emails
	Roles      []UserRole  `json:"roles,omitempty"`      // List of user roles
	ExternalID string      `json:"externalId,omitempty"` // External identifier from SAML provider, maxLength: 5000
	Active     bool        `json:"active"`               // Whether the user is active
}

// UpdateUserParams represents parameters for updating a user (PUT)
type UpdateUserParams = CreateUserParams

// PatchOperation represents a single operation in a PATCH request
type PatchOperation struct {
	Op    string `json:"op"`              // Operation type: "add", "remove", or "replace"
	Path  string `json:"path,omitempty"`  // Path to the value in the resource
	Value any    `json:"value,omitempty"` // Value to apply (can be object, array, or string)
}

// PatchUserParams represents parameters for patching a user
type PatchUserParams struct {
	Operations []PatchOperation `json:"Operations"` // Set of operations to perform
}

// ListUsersParams represents query parameters for listing users
type ListUsersParams struct {
	StartIndex int `json:"startIndex,omitempty"` // 1-based index of first result
	Count      int `json:"count,omitempty"`      // Maximum number of results per page
}

// ListUsersResponse represents the response from listing users
type ListUsersResponse struct {
	Schemas      []string `json:"schemas"`      // SCIM schemas used
	TotalResults int      `json:"totalResults"` // Total number of results
	StartIndex   int      `json:"startIndex"`   // Starting index of this page
	ItemsPerPage int      `json:"itemsPerPage"` // Number of items in this page
	Resources    []User   `json:"Resources"`    // List of users
}

// scimOptions returns RequestOptions for SCIM content type
var scimOptions = &utils.RequestOptions{
	ContentType: "application/scim+json",
	Accept:      "application/scim+json",
}

// CreateUser creates a new user in the organization
func CreateUser(client *utils.ProofClient, organizationID string, params *CreateUserParams) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Users/"
	return client.Post(path, params, scimOptions)
}

// GetUser retrieves a specific user by ID
func GetUser(client *utils.ProofClient, organizationID, userID string) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Users/" + userID
	return client.Get(path, scimOptions)
}

// ListUsers retrieves a list of users with optional pagination
func ListUsers(client *utils.ProofClient, organizationID string, params *ListUsersParams) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Users"

	// Build query parameters
	if params != nil {
		queryParams := url.Values{}
		if params.StartIndex > 0 {
			queryParams.Set("startIndex", fmt.Sprintf("%d", params.StartIndex))
		}
		if params.Count > 0 {
			queryParams.Set("count", fmt.Sprintf("%d", params.Count))
		}
		if len(queryParams) > 0 {
			path += "?" + queryParams.Encode()
		}
	}

	return client.Get(path, scimOptions)
}

// UpdateUser replaces all fields of a user (PUT)
func UpdateUser(client *utils.ProofClient, organizationID, userID string, params *UpdateUserParams) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Users/" + userID
	return client.Put(path, params, scimOptions)
}

// PatchUser performs partial updates on a user
func PatchUser(client *utils.ProofClient, organizationID, userID string, params *PatchUserParams) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Users/" + userID
	return client.Patch(path, params, scimOptions)
}

// DeleteUser deletes a user from the organization
func DeleteUser(client *utils.ProofClient, organizationID, userID string) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Users/" + userID
	return client.Delete(path, scimOptions)
}
