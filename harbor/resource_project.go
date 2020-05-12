package harbor

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/benosman/terraform-provider-harbor/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var pathProjects = "/api/v2.0/projects"

type ProjectCreateOpts struct {
	ProjectName string   `json:"project_name"`
	Metadata    metadata `json:"metadata"`
}

type ProjectOpts struct {
	Name        string   `json:"name"`
	Metadata    metadata `json:"metadata"`
}

type metadata struct {
	AutoScan string `json:"auto_scan"`
	Public   string `json:"public"`
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"auto_scan": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
		},
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceProjectSetToState(d *schema.ResourceData, project ProjectOpts) {
	d.Set("name", project.Name)
	d.Set("public", project.Metadata.Public)
	d.Set("auto_scan", project.Metadata.AutoScan)
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	body := ProjectCreateOpts{
		ProjectName: d.Get("name").(string),
		Metadata: metadata{
			AutoScan: strconv.FormatBool(d.Get("vulnerability_scanning").(bool)),
			Public:   strconv.FormatBool(d.Get("public").(bool)),
		},
	}

	var resp *client.ClientResponse
	resp, err := apiClient.SendRequestFull("POST", pathProjects, body, 0)
	if err != nil {
		return fmt.Errorf("[ERROR] Unable to create project: %s", err)
	}

	location := resp.Headers.Get("location")
	_, projectId := path.Split(location)
	d.SetId(projectId)

	return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	resp, err := apiClient.SendRequest("GET", pathProjects + "/" + d.Id(), nil, 0)
	if err != nil {
		return err
	}

	var jsonProject ProjectOpts

	err = json.Unmarshal([]byte(resp), &jsonProject)

	/*if err == nil {
		return fmt.Errorf("[RESPONSE] %s", resp)
	}*/
	if err != nil {
		return fmt.Errorf("[ERROR] Unable to unmarshal: %s", err)
	}

	resourceProjectSetToState(d, jsonProject)

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)

	body := ProjectOpts{
		Name: d.Get("name").(string),
		Metadata: metadata{
			AutoScan: strconv.FormatBool(d.Get("auto_scan").(bool)),
			Public:   strconv.FormatBool(d.Get("public").(bool)),
		},
	}

	_, err := apiClient.SendRequest("PUT", pathProjects + "/" + d.Id(), body, 0)
	if err != nil {
		return err
	}

	return resourceProjectRead(d, m)
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	_, err := apiClient.SendRequest("DELETE", pathProjects + "/" + d.Id(), nil, 0)
	return err
}
