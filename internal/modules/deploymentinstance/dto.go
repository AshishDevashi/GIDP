package deploymentinstance

type Response struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	EC2InstanceID   string `json:"ec2_instance_id,omitempty"`
	PublicIP        string `json:"public_ip,omitempty"`
	PrivateIP       string `json:"private_ip,omitempty"`
	APIServerURL    string `json:"api_server_url,omitempty"`
	AuthType        string `json:"auth_type"`
	CredentialsRef  string `json:"credentials_ref,omitempty"`
	MaxDeployments  int32  `json:"max_deployments"`
	Status          string `json:"status"`
	LastError       string `json:"last_error,omitempty"`
	Workspace       string `json:"workspace"`
	SSHKeyName      string `json:"ssh_key_name"`
	SecurityGroupID string `json:"security_group_id,omitempty"`
	CreatedBy       string `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}
