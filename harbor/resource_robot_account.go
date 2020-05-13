package harbor

import (
	"encoding/json"
	"fmt"
	"github.com/benosman/terraform-provider-harbor/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"path"
)

var pathRobot string = "/api/v2.0/projects"

type robot struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Access      []access `json:"access"`
	ExpiresAt   int      `json:"expires_at"`
}

type access struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

type robotAccount struct {
	Token     string `json:"token,omitempty"`
	RobotID   int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	ExpiresAt int    `json:"expires_at,omitempty"`
}

func resourceRobotAccount() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
			},
			"allow_push": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"allow_helm_pull": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"allow_helm_push": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"never_expires": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"expires_at": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ComputedWhen: []string{
					"never_expire",
				},
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
		Create: resourceRobotAccountCreate,
		Read:   resourceRobotAccountRead,
		Update: resourceRobotAccountUpdate,
		Delete: resourceRobotAccountDelete,
	}
}

func resourceRobotSetToState(d *schema.ResourceData, robot robotAccount) {
	d.Set("expires_at", robot.ExpiresAt)
}

func resourceRobotAccountCreate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	projectid := d.Get("project_id").(string)

	url := pathRobot + "/" + projectid + "/robots"

	resourceBase := "/project/" + projectid
	resourceRepo := resourceBase + "/repository"
	resourceHelmPull := resourceBase + "/helm-chart"
	resourceHelmPush := resourceBase + "/helm-chart-version"

	var accessList []access

	if d.Get("allow_push").(bool) {
		accessList = append(accessList, access{
			Action: "push",
			Resource: resourceRepo,
		})
	}

	if d.Get("allow_helm_pull").(bool) {
		accessList = append(accessList, access{
			Action: "read",
			Resource: resourceHelmPull,
		})
	}

	if d.Get("allow_helm_push").(bool) {
		accessList = append(accessList, access{
			Action: "create",
			Resource: resourceHelmPush,
		})
	}

	if d.Get("never_expires").(bool) {
		_ = d.Set("expires_at", -1)
	}

	body := robot{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Access:      accessList,
		ExpiresAt:   d.Get("expires_at").(int),
	}
	/*
	{"access":[{"action":"push","resource":"/project/7/repository"},{"action":"read","resource":"/project/7/helm-chart"},{"action":"create","resource":"/project/7/helm-chart-version"}],"expires_at":-1,"name":"drone"}
	 */
	resp, err := apiClient.SendRequestFull("POST", url, body, 201)
	if err != nil {
		return err
	}

	var jsonData robotAccount

	err = json.Unmarshal([]byte(resp.Body), &jsonData)
	if err != nil {
		return fmt.Errorf("[ERROR] Unable to unmarshal: %s", err)
	}

	location := resp.Headers.Get("location")
	_, robotId := path.Split(location)
	d.SetId(robotId)

	d.Set("token", jsonData.Token)
	return resourceRobotAccountRead(d, m)
}

func resourceRobotAccountRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	projectid := d.Get("project_id").(string)
	url := pathRobot + "/" + projectid + "/robots/" + d.Id()

	resp, err := apiClient.SendRequest("GET", url, nil, 200)
	if err != nil {
		return err
	}

	var jsonData robotAccount

	err = json.Unmarshal([]byte(resp), &jsonData)
	if err != nil {
		return fmt.Errorf("[ERROR] Unable to unmarshal: %s", err)
	}

	resourceRobotSetToState(d, jsonData)

	return nil
}

func resourceRobotAccountUpdate(d *schema.ResourceData, m interface{}) error {
	// apiClient := m.(*client.Client)

	return resourceRobotAccountRead(d, m)
}

func resourceRobotAccountDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*client.Client)
	projectid := d.Get("project_id").(string)
	url := pathRobot + "/" + projectid + "/robots/" + d.Id()
	_, err := apiClient.SendRequest("DELETE", url, nil, 0)
	return err
}
