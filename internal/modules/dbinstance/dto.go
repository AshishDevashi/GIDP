package dbinstance

// Response is the public representation of a persisted DB instance.
type Response struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	Engine             string `json:"engine"`
	EngineVersion      string `json:"engine_version"`
	Provider           string `json:"provider"`
	Region             string `json:"region"`
	InstanceType       string `json:"instance_type"`
	StorageGB          int32  `json:"storage_gb"`
	Status             string `json:"status"`
	ContainerStatus    string `json:"container_status"`
	StatusMessage      string `json:"status_message,omitempty"`
	Workspace          string `json:"workspace"`
	SSHKeyName         string `json:"ssh_key_name"`
	AdminUsername      string `json:"admin_username"`
	AdminSecretName    string `json:"admin_secret_name,omitempty"`
	PostgresPort       int32  `json:"postgres_port"`
	PostgresImage      string `json:"postgres_image,omitempty"`
	ProviderInstanceID string `json:"provider_instance_id,omitempty"`
	AvailabilityZone   string `json:"availability_zone,omitempty"`
	PublicIP           string `json:"public_ip,omitempty"`
	PrivateIP          string `json:"private_ip,omitempty"`
	SecurityGroupID    string `json:"security_group_id,omitempty"`
	VolumeID           string `json:"volume_id,omitempty"`
	CreatedBy          string `json:"created_by,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
