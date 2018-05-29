package iam

import (
	"encoding/json"
	"errors"

	"github.com/jeffail/gabs"
)

const permissionAPIVersion = "1"

type PermissionsService struct {
	client *Client
}

type GetPermissionOptions struct {
	ID     *string `url:"_id,omitempty"`
	Name   *string `url:"name,omitempty"`
	RoleID *string `url:"roleId,omitempty"`
}

func (p *PermissionsService) GetPermissionByID(id string) (*Permission, *Response, error) {
	return p.GetPermission(&GetPermissionOptions{ID: &id}, nil)
}

func (p *PermissionsService) GetPermissionByName(name string) (*Permission, *Response, error) {
	return p.GetPermission(&GetPermissionOptions{Name: &name}, nil)
}

func (p *PermissionsService) GetPermissionsByRoleID(roleID string) (*[]Permission, *Response, error) {
	opt := &GetPermissionOptions{
		RoleID: &roleID,
	}
	req, err := p.client.NewIDMRequest("GET", "authorize/identity/Permission", opt, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("api-version", permissionAPIVersion)

	var responseStruct struct {
		Total int          `json:"total"`
		Entry []Permission `json:"entry"`
	}

	resp, err := p.client.DoSigned(req, &responseStruct)
	if err != nil {
		return nil, resp, err
	}
	return &responseStruct.Entry, resp, err
}

func (p *PermissionsService) GetPermission(opt *GetPermissionOptions, options ...OptionFunc) (*Permission, *Response, error) {
	req, err := p.client.NewIDMRequest("GET", "authorize/identity/Permission", opt, options)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("api-version", permissionAPIVersion)

	var bundleResponse interface{}

	resp, err := p.client.DoSigned(req, &bundleResponse)
	if err != nil {
		return nil, resp, err
	}
	permissions, err := p.parseFromBundle(bundleResponse)
	if err != nil {
		return nil, resp, err
	}
	return &(*permissions)[0], resp, nil
}

func (p *PermissionsService) GetPermissions(opt GetPermissionOptions, options ...OptionFunc) (*[]Permission, *Response, error) {
	req, err := p.client.NewIDMRequest("GET", "authorize/identity/Permission", opt, options)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("api-version", permissionAPIVersion)

	var bundleResponse interface{}

	resp, err := p.client.DoSigned(req, &bundleResponse)
	if err != nil {
		return nil, resp, err
	}
	permissions, err := p.parseFromBundle(bundleResponse)
	return permissions, resp, err
}

func (p *PermissionsService) parseFromBundle(bundleResponse interface{}) (*[]Permission, error) {
	m, _ := json.Marshal(bundleResponse)
	jsonParsed, err := gabs.ParseJSON(m)
	if err != nil {
		return nil, err
	}
	count, ok := jsonParsed.S("total").Data().(float64)
	if !ok || count == 0 {
		return nil, errors.New("empty result")
	}
	permissions := make([]Permission, int64(count))

	children, _ := jsonParsed.S("entry").Children()
	for i, r := range children {
		var p Permission
		p.ID = r.Path("id").Data().(string)
		p.Category, _ = r.Path("category").Data().(string)
		p.Name, _ = r.Path("name").Data().(string)
		p.Description, _ = r.Path("description").Data().(string)
		p.Type, _ = r.Path("type").Data().(string)
		permissions[i] = p
	}
	return &permissions, nil
}
